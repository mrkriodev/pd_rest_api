package services

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"pdrest/internal/data"
	"pdrest/internal/domain"
	"time"
)

type RouletteService struct {
	repo      data.RouletteRepository
	userRepo  data.UserRepository
	prizeRepo data.PrizeRepository
	eventRepo data.EventRepository
}

func NewRouletteService(r data.RouletteRepository, userRepo data.UserRepository, prizeRepo data.PrizeRepository, eventRepo data.EventRepository) *RouletteService {
	return &RouletteService{
		repo:      r,
		userRepo:  userRepo,
		prizeRepo: prizeRepo,
		eventRepo: eventRepo,
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
	var preauthToken *domain.RoulettePreauthToken
	var err error
	var wasUnregistered bool

	// Get session_id and IP from context
	sessionID, hasSessionID := ctx.Value("session_id").(string)
	ipAddress, hasIPAddress := ctx.Value("ip_address").(string)

	// If preauth_token is not provided, generate it from session_id + IP
	if req.PreauthToken == "" {
		if !hasSessionID || !hasIPAddress {
			return nil, errors.New("preauth_token is required, or X-SESSION-ID and IP address must be provided")
		}

		// Generate token from session_id + IP (same logic as GetPreauthToken)
		token := generateTokenFromSessionAndIP(sessionID, ipAddress)

		// Get or create preauth token
		preauthToken, err = s.repo.GetPreauthToken(ctx, token)
		if err != nil {
			return nil, fmt.Errorf("failed to get preauth token: %w", err)
		}

		// If token doesn't exist, create it
		if preauthToken == nil {
			// Get active on_start config (no event_id)
			config, err := s.repo.GetRouletteConfigByType(ctx, domain.RouletteTypeOnStart, nil)
			if err != nil {
				return nil, fmt.Errorf("failed to get roulette config: %w", err)
			}
			if config == nil || !config.IsActive {
				return nil, errors.New("roulette config not found or inactive")
			}

			// Create new preauth token (no user_uuid, expires far in the future - 10 years)
			expiresAt := time.Now().Add(10 * 365 * 24 * time.Hour).UnixMilli()
			preauthToken = &domain.RoulettePreauthToken{
				Token:            token,
				UserUUID:         nil, // Always nil for unauthenticated users
				RouletteConfigID: config.ID,
				IsUsed:           false,
				ExpiresAt:        expiresAt,
			}

			if err := s.repo.CreatePreauthToken(ctx, preauthToken); err != nil {
				return nil, fmt.Errorf("failed to create preauth token: %w", err)
			}
		}

		// Check if user was unregistered (no user_uuid linked)
		wasUnregistered = preauthToken.UserUUID == nil

		// Create user if session_id and IP are provided but user doesn't exist
		if s.userRepo != nil && hasSessionID && hasIPAddress {
			if err := s.userRepo.CreateOrUpdateUserBySession(sessionID, ipAddress); err != nil {
				// Log error but don't fail the request
				_ = err
			}
		}
	} else {
		// Validate provided preauth token
		preauthToken, err = s.repo.ValidatePreauthToken(ctx, req.PreauthToken)
		if err != nil {
			return nil, fmt.Errorf("invalid preauth token: %w", err)
		}

		// Check if user was unregistered (no user_uuid linked)
		wasUnregistered = preauthToken.UserUUID == nil

		// Create user if session_id and IP are provided but user doesn't exist
		if s.userRepo != nil && hasSessionID && hasIPAddress {
			if err := s.userRepo.CreateOrUpdateUserBySession(sessionID, ipAddress); err != nil {
				// Log error but don't fail the request
				_ = err
			}
		}
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
			response := &domain.TakePrizeResponse{
				Success: true,
				Prize:   *roulette.Prize,
				Message: "Prize already taken",
			}
			// Return preauth_token if user was unregistered
			if wasUnregistered {
				response.PreauthToken = preauthToken.Token
			}
			return response, nil
		}
		return nil, errors.New("prize already taken but no prize found")
	}

	// Check if user has completed all spins
	if roulette.SpinNumber < config.MaxSpins {
		return nil, fmt.Errorf("must complete all %d spins before taking prize", config.MaxSpins)
	}

	// Determine prize based on event or default
	prizeValue, prizeType, err := s.determinePrize(ctx, config, roulette)
	if err != nil {
		return nil, fmt.Errorf("failed to determine prize: %w", err)
	}

	// Create prize record
	now := time.Now().UnixMilli()
	prize := &domain.Prize{
		EventID:        config.EventID,
		PreauthTokenID: &preauthToken.ID,
		RouletteID:     &roulette.ID,
		PrizeValue:     prizeValue,
		PrizeType:      prizeType,
		AwardedAt:      now,
		CreatedAt:      now,
	}

	// If user is authenticated, also set user_uuid
	if preauthToken.UserUUID != nil {
		prize.UserID = preauthToken.UserUUID
	}

	// Create prize in database
	if s.prizeRepo != nil {
		if err := s.prizeRepo.CreatePrize(ctx, prize); err != nil {
			return nil, fmt.Errorf("failed to create prize record: %w", err)
		}
	}

	// Update roulette with prize
	if err := s.repo.TakePrize(ctx, roulette.ID, prizeValue); err != nil {
		return nil, fmt.Errorf("failed to take prize: %w", err)
	}

	// Mark preauth token as used
	if err := s.repo.MarkPreauthTokenAsUsed(ctx, preauthToken.ID); err != nil {
		// Log error but don't fail the request
		_ = err
	}

	response := &domain.TakePrizeResponse{
		Success: true,
		Prize:   prizeValue,
		Message: "Prize taken successfully",
	}

	// Return preauth_token if user was unregistered
	if wasUnregistered {
		response.PreauthToken = preauthToken.Token
	}

	return response, nil
}

// GetPreauthToken gets or creates a preauth token for on_start roulette based on session_id and IP
// Only for unauthenticated users. Returns existing token if it exists, otherwise creates a new one.
func (s *RouletteService) GetPreauthToken(ctx context.Context, sessionID, ipAddress string) (string, error) {
	// Generate deterministic token from session_id + IP
	token := generateTokenFromSessionAndIP(sessionID, ipAddress)

	// Check if token already exists
	existingToken, err := s.repo.GetPreauthToken(ctx, token)
	if err != nil {
		return "", fmt.Errorf("failed to check existing token: %w", err)
	}

	// If token exists, return it
	if existingToken != nil {
		return existingToken.Token, nil
	}

	// Get active on_start config (no event_id)
	config, err := s.repo.GetRouletteConfigByType(ctx, domain.RouletteTypeOnStart, nil)
	if err != nil {
		return "", fmt.Errorf("failed to get roulette config: %w", err)
	}
	if config == nil || !config.IsActive {
		return "", errors.New("roulette config not found or inactive")
	}

	// Create new preauth token (no user_uuid, expires far in the future - 10 years)
	expiresAt := time.Now().Add(10 * 365 * 24 * time.Hour).UnixMilli()
	preauthToken := &domain.RoulettePreauthToken{
		Token:            token,
		UserUUID:         nil, // Always nil for unauthenticated users
		RouletteConfigID: config.ID,
		IsUsed:           false,
		ExpiresAt:        expiresAt,
	}

	if err := s.repo.CreatePreauthToken(ctx, preauthToken); err != nil {
		return "", fmt.Errorf("failed to create preauth token: %w", err)
	}

	return token, nil
}

// LinkPreauthTokenToUser links a preauth token to a user UUID (called after successful auth)
func (s *RouletteService) LinkPreauthTokenToUser(ctx context.Context, preauthToken string, userUUID string) error {
	return s.repo.UpdatePreauthTokenUserUUID(ctx, preauthToken, userUUID)
}

// determinePrize determines the prize value based on event or default
func (s *RouletteService) determinePrize(ctx context.Context, config *domain.RouletteConfig, roulette *domain.Roulette) (string, domain.PrizeType, error) {
	// If this is a during_event roulette and event_id is set, get event rewards
	if config.Type == domain.RouletteTypeDuringEvent && config.EventID != nil {
		if s.eventRepo == nil {
			return "Default Prize", domain.PrizeTypeRouletteDuringEvent, nil
		}

		// Get event by ID
		event, err := s.eventRepo.GetEventByID(ctx, *config.EventID)
		if err != nil {
			return "Default Prize", domain.PrizeTypeRouletteDuringEvent, nil
		}

		// If event found and has rewards, use the first reward value
		if event != nil && len(event.Reward) > 0 {
			// Use the first reward as the prize (or implement more complex logic)
			prizeValue := event.Reward[0].Value
			return prizeValue, domain.PrizeTypeRouletteDuringEvent, nil
		}

		// Event found but no rewards, use default
		return "Default Prize", domain.PrizeTypeRouletteDuringEvent, nil
	}

	// For on_start roulette, use default prize
	// You can implement more complex logic here based on spin results, config, etc.
	return "Default Prize", domain.PrizeTypeRouletteOnStart, nil
}

// generateTokenFromSessionAndIP generates a deterministic token from session_id and IP address
func generateTokenFromSessionAndIP(sessionID, ipAddress string) string {
	// Combine session_id and IP
	data := fmt.Sprintf("%s:%s", sessionID, ipAddress)

	// Generate SHA256 hash
	hash := sha256.Sum256([]byte(data))

	// Convert to hex string (64 characters)
	return hex.EncodeToString(hash[:])
}

// CreatePreauthToken creates a preauth token (typically called from browser)
// DEPRECATED: Use GetPreauthToken instead for on_start roulette
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
