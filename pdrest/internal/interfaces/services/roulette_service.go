package services

import (
	"context"
	"errors"
	"fmt"
	"pdrest/internal/data"
	"pdrest/internal/domain"
)

type RouletteService struct {
	repo     data.RouletteRepository
	userRepo data.UserRepository
}

func NewRouletteService(r data.RouletteRepository, userRepo data.UserRepository) *RouletteService {
	return &RouletteService{
		repo:     r,
		userRepo: userRepo,
	}
}

// GetRouletteStatus gets the current status of roulette by preauth token
func (s *RouletteService) GetRouletteStatus(ctx context.Context, preauthToken string) (*domain.GetRouletteStatusResponse, error) {
	// Validate and get preauth token
	token, err := s.repo.ValidatePreauthToken(ctx, preauthToken)
	if err != nil {
		return nil, fmt.Errorf("invalid preauth token: %w", err)
	}

	// Get config
	config, err := s.repo.GetRouletteConfigByID(ctx, token.RouletteConfigID)
	if err != nil {
		return nil, fmt.Errorf("failed to get roulette config: %w", err)
	}
	if config == nil {
		return &domain.GetRouletteStatusResponse{
			CanSpin: false,
		}, nil
	}

	// Get roulette by preauth token
	roulette, err := s.repo.GetRouletteByPreauthToken(ctx, token.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get roulette: %w", err)
	}

	response := &domain.GetRouletteStatusResponse{
		Config:     config,
		CanSpin:    false,
		PrizeTaken: false,
	}

	if roulette == nil {
		// User hasn't started yet
		response.RemainingSpins = config.MaxSpins
		response.CanSpin = true
		return response, nil
	}

	response.Roulette = roulette
	response.PrizeTaken = roulette.PrizeTaken

	if roulette.PrizeTaken {
		// Prize already taken, no more spins
		response.RemainingSpins = 0
		response.CanSpin = false
	} else {
		// Calculate remaining spins
		response.RemainingSpins = config.MaxSpins - roulette.SpinNumber
		response.CanSpin = response.RemainingSpins > 0
	}

	return response, nil
}

// Spin performs a spin using preauth token
func (s *RouletteService) Spin(ctx context.Context, req *domain.SpinRequest) (*domain.SpinResponse, error) {
	// Validate preauth token
	preauthToken, err := s.repo.ValidatePreauthToken(ctx, req.PreauthToken)
	if err != nil {
		return nil, fmt.Errorf("invalid preauth token: %w", err)
	}

	// Get config
	config, err := s.repo.GetRouletteConfigByID(ctx, preauthToken.RouletteConfigID)
	if err != nil {
		return nil, fmt.Errorf("failed to get roulette config: %w", err)
	}
	if config == nil || !config.IsActive {
		return nil, errors.New("roulette config not found or inactive")
	}

	// Get or create roulette
	roulette, err := s.repo.GetRouletteByPreauthToken(ctx, preauthToken.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get roulette: %w", err)
	}

	// Check if prize already taken
	if roulette != nil && roulette.PrizeTaken {
		return nil, errors.New("prize already taken, no more spins available")
	}

	// Check if user has remaining spins
	if roulette != nil {
		if roulette.SpinNumber >= config.MaxSpins {
			return nil, errors.New("maximum spins reached")
		}
	}

	// Perform spin logic (increment spin number)
	if roulette == nil {
		// Create new roulette entry
		roulette = &domain.Roulette{
			RouletteConfigID: config.ID,
			PreauthTokenID:   preauthToken.ID,
			SpinNumber:       1,
			PrizeTaken:       false,
			SpinResult:       make(map[string]interface{}),
		}
		if err := s.repo.CreateRoulette(ctx, roulette); err != nil {
			return nil, fmt.Errorf("failed to create roulette: %w", err)
		}
	} else {
		// Update existing roulette
		roulette.SpinNumber++
		if err := s.repo.UpdateRoulette(ctx, roulette); err != nil {
			return nil, fmt.Errorf("failed to update roulette: %w", err)
		}
	}

	// Mark preauth token as used
	if err := s.repo.MarkPreauthTokenAsUsed(ctx, preauthToken.ID); err != nil {
		return nil, fmt.Errorf("failed to mark preauth token as used: %w", err)
	}

	// Calculate remaining spins
	remainingSpins := config.MaxSpins - roulette.SpinNumber
	canSpin := remainingSpins > 0 && !roulette.PrizeTaken

	return &domain.SpinResponse{
		Roulette:       roulette,
		RemainingSpins: remainingSpins,
		CanSpin:        canSpin,
	}, nil
}

// TakePrize allows user to take the prize after completing all spins
func (s *RouletteService) TakePrize(ctx context.Context, req *domain.TakePrizeRequest) (*domain.TakePrizeResponse, error) {
	// Validate preauth token
	preauthToken, err := s.repo.ValidatePreauthToken(ctx, req.PreauthToken)
	if err != nil {
		return nil, fmt.Errorf("invalid preauth token: %w", err)
	}

	// Get config
	config, err := s.repo.GetRouletteConfigByID(ctx, preauthToken.RouletteConfigID)
	if err != nil {
		return nil, fmt.Errorf("failed to get roulette config: %w", err)
	}
	if config == nil || !config.IsActive {
		return nil, errors.New("roulette config not found or inactive")
	}

	// Get roulette
	roulette, err := s.repo.GetRouletteByPreauthToken(ctx, preauthToken.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get roulette: %w", err)
	}
	if roulette == nil {
		return nil, errors.New("roulette not found, must spin first")
	}

	// Check if prize already taken
	if roulette.PrizeTaken {
		if roulette.Prize != nil {
			return &domain.TakePrizeResponse{
				Success: true,
				Prize:   *roulette.Prize,
				Message: "Prize already taken",
			}, nil
		}
		return nil, errors.New("prize already taken but no prize found")
	}

	// Check if user has completed all spins
	if roulette.SpinNumber < config.MaxSpins {
		return nil, fmt.Errorf("must complete all %d spins before taking prize", config.MaxSpins)
	}

	// Determine prize (this is a placeholder - implement your prize logic here)
	// For now, we'll use a simple prize based on the spin result or config
	prize := "Default Prize" // TODO: Implement actual prize determination logic

	// Take prize
	if err := s.repo.TakePrize(ctx, roulette.ID, prize); err != nil {
		return nil, fmt.Errorf("failed to take prize: %w", err)
	}

	// Mark preauth token as used
	if err := s.repo.MarkPreauthTokenAsUsed(ctx, preauthToken.ID); err != nil {
		// Log error but don't fail the request
		_ = err
	}

	// Internal user registration with session_id and IP (silent, no return info)
	// Extract session_id and IP from context if available
	if s.userRepo != nil {
		if sessionID, ok := ctx.Value("session_id").(string); ok {
			if ipAddress, ok := ctx.Value("ip_address").(string); ok {
				// Register user internally (ignore errors - this is a background operation)
				_ = s.userRepo.CreateOrUpdateUserBySession(sessionID, ipAddress)
			}
		}
	}

	return &domain.TakePrizeResponse{
		Success: true,
		Prize:   prize,
		Message: "Prize taken successfully",
	}, nil
}

// CreatePreauthToken creates a preauth token (typically called from browser)
func (s *RouletteService) CreatePreauthToken(ctx context.Context, rouletteType domain.RouletteType, eventID *string, token string, expiresAt int64, userUUID *string) error {
	// Get active config
	config, err := s.repo.GetRouletteConfigByType(ctx, rouletteType, eventID)
	if err != nil {
		return fmt.Errorf("failed to get roulette config: %w", err)
	}
	if config == nil || !config.IsActive {
		return errors.New("roulette config not found or inactive")
	}

	// Create preauth token
	preauthToken := &domain.RoulettePreauthToken{
		Token:            token,
		UserUUID:         userUUID,
		RouletteConfigID: config.ID,
		IsUsed:           false,
		ExpiresAt:        expiresAt,
	}

	if err := s.repo.CreatePreauthToken(ctx, preauthToken); err != nil {
		return fmt.Errorf("failed to create preauth token: %w", err)
	}

	return nil
}
