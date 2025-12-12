package services

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"math/rand"
	"pdrest/internal/data"
	"pdrest/internal/domain"
	"strings"
	"time"
)

type RouletteService struct {
	repo           data.RouletteRepository
	userRepo       data.UserRepository
	prizeRepo      data.PrizeRepository
	prizeValueRepo data.PrizeValueRepository
	eventRepo      data.EventRepository
}

type ContextKey string

const (
	ContextKeyAuthHeader ContextKey = "auth_header"
	ContextKeySessionID  ContextKey = "session_id"
	ContextKeyIPAddress  ContextKey = "ip_address"
)

func NewRouletteService(r data.RouletteRepository, userRepo data.UserRepository, prizeRepo data.PrizeRepository, prizeValueRepo data.PrizeValueRepository, eventRepo data.EventRepository) *RouletteService {
	return &RouletteService{
		repo:           r,
		userRepo:       userRepo,
		prizeRepo:      prizeRepo,
		prizeValueRepo: prizeValueRepo,
		eventRepo:      eventRepo,
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

// GetRouletteConfigByID returns roulette config by id
func (s *RouletteService) GetRouletteConfigByID(ctx context.Context, id int) (*domain.RouletteConfig, error) {
	if id <= 0 {
		return nil, errors.New("invalid roulette id")
	}
	return s.repo.GetRouletteConfigByID(ctx, id)
}

// Spin performs a spin using preauth token
func (s *RouletteService) Spin(ctx context.Context, preauthTokenStr string, req *domain.SpinRequest) (*domain.SpinResponse, error) {
	var preauthToken *domain.RoulettePreauthToken
	var err error

	// Get session_id and IP from context
	sessionID, hasSessionID := ctx.Value(ContextKeySessionID).(string)
	ipAddress, hasIPAddress := ctx.Value(ContextKeyIPAddress).(string)

	// If preauth_token is not provided, generate it from session_id + IP
	if preauthTokenStr == "" {
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
			// Get active on_start config (references startup event)
			config, err := s.repo.GetRouletteConfigByType(ctx, domain.RouletteTypeOnStart, "startup")
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
	} else {
		// Validate provided preauth token
		preauthToken, err = s.repo.ValidatePreauthToken(ctx, preauthTokenStr)
		if err != nil {
			return nil, fmt.Errorf("invalid preauth token: %w", err)
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

	// Validate roulette_id matches config
	if req.RouletteID != 0 && req.RouletteID != config.ID {
		return nil, errors.New("invalid roulette_id for provided preauth_token")
	}

	// If roulette is during_event, Authorization header is required
	if config.Type == domain.RouletteTypeDuringEvent {
		authHeader, _ := ctx.Value(ContextKeyAuthHeader).(string)
		if strings.TrimSpace(authHeader) == "" {
			return nil, errors.New("authorization is required for event roulette")
		}
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

	// Use event ID from config (always set, references startup for on_start, specific event for during_event)
	eventID := config.EventID

	// Get prize values for the event
	if s.prizeValueRepo == nil {
		return nil, errors.New("prize value repository is not initialized")
	}

	prizeValues, err := s.prizeValueRepo.GetPrizeValuesByEventID(ctx, eventID)
	if err != nil {
		return nil, fmt.Errorf("failed to get prize values: %w", err)
	}

	if len(prizeValues) == 0 {
		return nil, fmt.Errorf("no prize values configured for event: %s", eventID)
	}

	// Randomly select one prize value
	rand.Seed(time.Now().UnixNano())
	selectedPrizeValue := &prizeValues[rand.Intn(len(prizeValues))]

	// Store selected prize in spin_result
	if roulette.SpinResult == nil {
		roulette.SpinResult = make(map[string]interface{})
	}
	roulette.SpinResult["prize_value_id"] = selectedPrizeValue.ID
	roulette.SpinResult["prize_value"] = selectedPrizeValue.Value // Now int64 (points)
	roulette.SpinResult["prize_label"] = selectedPrizeValue.Label
	if selectedPrizeValue.SegmentID != nil {
		roulette.SpinResult["segment_id"] = *selectedPrizeValue.SegmentID
	}

	// Also store in prize field (will be used when taking prize)
	// Convert int64 to string for storage in Prize field (which is still string in domain)
	prizeValueStr := fmt.Sprintf("%d", selectedPrizeValue.Value)
	roulette.Prize = &prizeValueStr

	// Update roulette with selected prize
	if err := s.repo.UpdateRoulette(ctx, roulette); err != nil {
		return nil, fmt.Errorf("failed to update roulette with prize: %w", err)
	}

	// Calculate remaining spins
	remainingSpins := config.MaxSpins - roulette.SpinNumber
	if remainingSpins < 0 {
		remainingSpins = 0
	}

	// Build frontend-friendly response
	segmentID := "1"
	if selectedPrizeValue.SegmentID != nil {
		segmentID = *selectedPrizeValue.SegmentID
	}

	result := domain.SpinResult{
		SegmentID: segmentID,
		Label:     selectedPrizeValue.Label,
	}
	reward := domain.SpinReward{
		Type:   "eth",
		Amount: float64(selectedPrizeValue.Value) / 1e9, // Convert points to ETH (1 ETH = 10^9 points)
	}

	return &domain.SpinResponse{
		Result:    result,
		SpinsLeft: remainingSpins,
		Reward:    reward,
	}, nil
}

// TakePrize allows user to take the prize after completing all spins
func (s *RouletteService) TakePrize(ctx context.Context, preauthTokenStr string, req *domain.TakePrizeRequest) (*domain.TakePrizeResponse, error) {
	var preauthToken *domain.RoulettePreauthToken
	var err error
	var wasUnregistered bool

	// Get session_id and IP from context
	sessionID, hasSessionID := ctx.Value(ContextKeySessionID).(string)
	ipAddress, hasIPAddress := ctx.Value(ContextKeyIPAddress).(string)

	// If preauth_token is not provided, generate it from session_id + IP
	if preauthTokenStr == "" {
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
			// Get active on_start config (references startup event)
			config, err := s.repo.GetRouletteConfigByType(ctx, domain.RouletteTypeOnStart, "startup")
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
		preauthToken, err = s.repo.ValidatePreauthToken(ctx, preauthTokenStr)
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

	// Validate roulette_id matches config
	if req.RouletteID != 0 && req.RouletteID != config.ID {
		return nil, errors.New("invalid roulette_id for provided preauth_token")
	}

	// If roulette is during_event, Authorization header is required
	if config.Type == domain.RouletteTypeDuringEvent {
		authHeader, _ := ctx.Value(ContextKeyAuthHeader).(string)
		if strings.TrimSpace(authHeader) == "" {
			return nil, errors.New("authorization is required for event roulette")
		}
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

	// Get prize value from roulette (stored during spin)
	var prizeValue string
	var prizeValueID *int
	if roulette.Prize != nil && *roulette.Prize != "" {
		prizeValue = *roulette.Prize
	} else if roulette.SpinResult != nil {
		// Fallback to spin_result if prize field is not set
		// prize_value is now int64 (points), convert to string
		if pv, ok := roulette.SpinResult["prize_value"].(int64); ok {
			prizeValue = fmt.Sprintf("%d", pv)
		} else if pv, ok := roulette.SpinResult["prize_value"].(float64); ok {
			// Handle case where JSON unmarshaling converts int64 to float64
			prizeValue = fmt.Sprintf("%.0f", pv)
		}
		if pvID, ok := roulette.SpinResult["prize_value_id"].(int); ok {
			prizeValueID = &pvID
		}
	}

	// If no prize found, try to get from prize_value_id
	if prizeValue == "" && prizeValueID != nil && s.prizeValueRepo != nil {
		pv, err := s.prizeValueRepo.GetPrizeValueByID(ctx, *prizeValueID)
		if err == nil && pv != nil {
			prizeValue = fmt.Sprintf("%d", pv.Value)
		}
	}

	// If still no prize found, determine default prize
	if prizeValue == "" {
		var err error
		prizeValue, _, err = s.determinePrize(ctx, config, roulette)
		if err != nil {
			return nil, fmt.Errorf("failed to determine prize: %w", err)
		}
	}

	// Determine prize type
	prizeType := domain.PrizeTypeRouletteOnStart
	if config.Type == domain.RouletteTypeDuringEvent {
		prizeType = domain.PrizeTypeRouletteDuringEvent
	}

	// Get user UUID - must be set (user_uuid is now mandatory)
	var userID string
	if preauthToken.UserUUID != nil {
		userID = *preauthToken.UserUUID
	} else if hasSessionID && s.userRepo != nil {
		// Try to get user by session_id
		if user, err := s.userRepo.GetUserBySessionID(ctx, sessionID); err == nil && user != nil {
			userID = user.UserID
		} else {
			return nil, errors.New("user_uuid is required to take prize - user not found by session_id")
		}
	} else {
		return nil, errors.New("user_uuid is required to take prize")
	}

	// Create prize record
	now := time.Now().UnixMilli()
	eventID := config.EventID
	prize := &domain.Prize{
		EventID:        &eventID,
		UserID:         &userID,
		PrizeValueID:   prizeValueID,
		PreauthTokenID: &preauthToken.ID,
		RouletteID:     &roulette.ID,
		PrizeValue:     prizeValue,
		PrizeType:      prizeType,
		AwardedAt:      now,
		CreatedAt:      now,
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

	// Get active on_start config (references startup event)
	config, err := s.repo.GetRouletteConfigByType(ctx, domain.RouletteTypeOnStart, "startup")
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
	// If this is a during_event roulette, get event rewards
	if config.Type == domain.RouletteTypeDuringEvent {
		if s.eventRepo == nil {
			return "Default Prize", domain.PrizeTypeRouletteDuringEvent, nil
		}

		// Get event by ID
		event, err := s.eventRepo.GetEventByID(ctx, config.EventID)
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
func (s *RouletteService) CreatePreauthToken(ctx context.Context, rouletteType domain.RouletteType, eventID string, token string, expiresAt int64, userUUID *string) error {
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
