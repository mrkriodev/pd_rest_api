package services

import (
	"context"
	"errors"
	"fmt"
	"math"
	"pdrest/internal/data"
	"pdrest/internal/domain"
	"time"
)

type BetService struct {
	repo          data.BetRepository
	priceProvider *PriceProvider
	scheduler     *BetScheduler
}

func NewBetService(r data.BetRepository, priceProvider *PriceProvider, scheduler *BetScheduler) *BetService {
	return &BetService{
		repo:          r,
		priceProvider: priceProvider,
		scheduler:     scheduler,
	}
}

func (s *BetService) OpenBet(ctx context.Context, userUUID string, req *domain.OpenBetRequest) (*domain.OpenBetResponse, error) {
	// Validate side
	if req.Side != "pump" && req.Side != "dump" {
		return nil, errors.New("side must be 'pump' or 'dump'")
	}

	// Validate sum
	if req.Sum <= 0 {
		return nil, errors.New("sum must be greater than 0")
	}
	if req.Sum != math.Trunc(req.Sum) {
		return nil, errors.New("sum must be a whole number of USDT")
	}

	// Validate pair
	if req.Pair == "" {
		return nil, errors.New("pair is required")
	}

	// Validate timeframe
	if req.Timeframe <= 0 {
		return nil, errors.New("timeframe must be greater than 0")
	}

	// Validate open price
	if req.OpenPrice <= 0 {
		return nil, errors.New("openPrice must be greater than 0")
	}

	// Create bet
	bet := &domain.Bet{
		UserID:    userUUID,
		Side:      req.Side,
		Sum:       req.Sum,
		Pair:      req.Pair,
		Timeframe: req.Timeframe,
		OpenPrice: req.OpenPrice,
		OpenTime:  req.OpenTime,
	}

	if err := s.repo.CreateBet(ctx, bet); err != nil {
		return nil, fmt.Errorf("failed to create bet: %w", err)
	}

	// Schedule bet closing if scheduler is available
	// This will fetch price from Binance when bet opens, then schedule another fetch after timeframe
	if s.scheduler != nil {
		// Fetch current price from Binance when bet opens (optional, for verification)
		// The main price fetch happens at close time
		if err := s.scheduler.ScheduleBetClosing(bet.ID, bet.Pair, bet.OpenTime, bet.Timeframe); err != nil {
			// Log error but don't fail bet creation
			// The bet can still be closed manually via GetBetStatus
			_ = err
		}
	}

	return &domain.OpenBetResponse{
		ID: bet.ID,
	}, nil
}

func (s *BetService) GetBetStatus(ctx context.Context, betID int, userUUID string) (*domain.BetStatusResponse, error) {
	// Get bet from database
	bet, err := s.repo.GetBetByID(ctx, betID, userUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get bet: %w", err)
	}
	if bet == nil {
		return nil, errors.New("bet not found")
	}

	// Check if timeframe has passed (timeframe is in seconds)
	now := time.Now()
	timeframeDuration := time.Duration(bet.Timeframe) * time.Second
	expectedCloseTime := bet.OpenTime.Add(timeframeDuration)

	// If timeframe has passed and closePrice is not set, fetch from provider
	if now.After(expectedCloseTime) && bet.ClosePrice == nil {
		if s.priceProvider != nil {
			closePrice, err := s.priceProvider.GetPrice(bet.Pair)
			if err != nil {
				// Log error but don't fail - return current bet status without closePrice
				_ = err
			} else {
				// Update bet with closePrice and closeTime
				closeTime := expectedCloseTime
				if err := s.repo.UpdateBetClosePrice(ctx, betID, closePrice, closeTime); err != nil {
					// Log error but don't fail - return current bet status
					_ = err
				} else {
					// Update local bet object
					bet.ClosePrice = &closePrice
					bet.CloseTime = &closeTime
				}
			}
		}
	}

	return &domain.BetStatusResponse{
		Side:       bet.Side,
		Sum:        bet.Sum,
		Pair:       bet.Pair,
		Timeframe:  bet.Timeframe,
		OpenPrice:  bet.OpenPrice,
		ClosePrice: bet.ClosePrice,
	}, nil
}

func (s *BetService) GetUnfinishedBetsByUser(ctx context.Context, userUUID string) ([]domain.Bet, error) {
	bets, err := s.repo.GetUnfinishedBetsByUser(ctx, userUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get unfinished bets: %w", err)
	}

	return bets, nil
}
