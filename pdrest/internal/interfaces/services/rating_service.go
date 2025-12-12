package services

import (
	"context"
	"errors"
	"fmt"
	"pdrest/internal/data"
	"pdrest/internal/domain"
	"strconv"
	"strings"
	"sync"
)

type RatingService struct {
	repo      data.RatingRepository
	prizeRepo data.PrizeRepository
	betRepo   data.BetRepository
}

func NewRatingService(r data.RatingRepository, prizeRepo data.PrizeRepository, betRepo data.BetRepository) *RatingService {
	return &RatingService{
		repo:      r,
		prizeRepo: prizeRepo,
		betRepo:   betRepo,
	}
}

func (s *RatingService) GetUserAssets(ctx context.Context, userUUID string) (*domain.UserAssets, error) {
	if userUUID == "" {
		return nil, errors.New("user uuid is required")
	}

	if s.repo == nil {
		return nil, errors.New("rating repository is not configured")
	}

	// Collect and add prizes and winning bets to rating
	if err := s.collectPrizesAndBets(ctx, userUUID); err != nil {
		// Log error but don't fail the request
		_ = err
	}

	totals, err := s.repo.GetUserRatingTotals(ctx, userUUID)
	if err != nil {
		return nil, err
	}

	if totals == nil {
		totals = &domain.RatingTotals{}
	}

	return &domain.UserAssets{
		UserID:      userUUID,
		Points:      *totals,
		TotalPoints: totals.TotalPoints(),
	}, nil
}

// collectPrizesAndBets collects prizes and winning bets, converts them to points, and adds to rating
// This method is idempotent - it only processes prizes and bets created after the last rating entry
// Uses goroutines to fetch and process prizes and bets in parallel for better performance
func (s *RatingService) collectPrizesAndBets(ctx context.Context, userUUID string) error {
	if s.prizeRepo == nil || s.betRepo == nil || s.repo == nil {
		return nil // Skip if repositories not available
	}

	// Get max created_at from rating table to avoid processing previously added prizes/bets
	maxCreatedAt, err := s.repo.GetMaxCreatedAt(ctx, userUUID)
	if err != nil {
		return fmt.Errorf("failed to get max created_at: %w", err)
	}

	// Use channels and goroutines to fetch prizes and bets in parallel
	type prizeResult struct {
		prizes []domain.Prize
		err    error
	}
	type betResult struct {
		bets []domain.Bet
		err  error
	}

	prizeChan := make(chan prizeResult, 1)
	betChan := make(chan betResult, 1)

	// Fetch prizes in parallel
	go func() {
		prizes, err := s.prizeRepo.GetPrizesByUserID(ctx, userUUID)
		prizeChan <- prizeResult{prizes: prizes, err: err}
	}()

	// Fetch winning bets in parallel
	go func() {
		bets, err := s.betRepo.GetWinningBetsByUser(ctx, userUUID)
		betChan <- betResult{bets: bets, err: err}
	}()

	// Wait for both results
	prizeRes := <-prizeChan
	betRes := <-betChan

	if prizeRes.err != nil {
		return fmt.Errorf("failed to get prizes: %w", prizeRes.err)
	}
	if betRes.err != nil {
		return fmt.Errorf("failed to get winning bets: %w", betRes.err)
	}

	// Process prizes and bets in parallel using goroutines (map-reduce pattern)
	var wg sync.WaitGroup
	var prizePoints int64
	var betPoints int64
	var prizeErr, betErr error

	// Process prizes in parallel
	wg.Add(1)
	go func() {
		defer wg.Done()
		prizePoints, prizeErr = s.processPrizes(prizeRes.prizes, maxCreatedAt)
	}()

	// Process bets in parallel
	wg.Add(1)
	go func() {
		defer wg.Done()
		betPoints, betErr = s.processBets(betRes.bets, maxCreatedAt)
	}()

	// Wait for both processing to complete
	wg.Wait()

	if prizeErr != nil {
		return fmt.Errorf("failed to process prizes: %w", prizeErr)
	}
	if betErr != nil {
		return fmt.Errorf("failed to process bets: %w", betErr)
	}

	// Add points if there are any new prizes or bets
	if prizePoints > 0 || betPoints > 0 {
		totalPoints := prizePoints + betPoints
		description := fmt.Sprintf("Prizes and winning bets: %d points (prizes: %d, bets: %d)", totalPoints, prizePoints, betPoints)
		if err := s.repo.AddPoints(ctx, userUUID, totalPoints, domain.RatingSourceBetBonus, description); err != nil {
			return fmt.Errorf("failed to add prize and bet points: %w", err)
		}
	}

	return nil
}

// processPrizes calculates total points from prizes (only those awarded after maxCreatedAt)
func (s *RatingService) processPrizes(prizes []domain.Prize, maxCreatedAt *int64) (int64, error) {
	totalPoints := int64(0)
	for _, prize := range prizes {
		// Skip prizes that were awarded before the last rating entry
		if maxCreatedAt != nil && prize.AwardedAt <= *maxCreatedAt {
			continue
		}

		points, err := s.parsePrizeValueToPoints(prize.PrizeValue)
		if err != nil {
			// Skip invalid prize values
			continue
		}
		totalPoints += points
	}
	return totalPoints, nil
}

// processBets calculates total points from winning bets (only those closed after maxCreatedAt)
func (s *RatingService) processBets(bets []domain.Bet, maxCreatedAt *int64) (int64, error) {
	totalPoints := int64(0)
	for _, bet := range bets {
		// Skip bets that were closed before the last rating entry
		if maxCreatedAt != nil && bet.CloseTime != nil {
			closeTimeMs := bet.CloseTime.UnixMilli()
			if closeTimeMs <= *maxCreatedAt {
				continue
			}
		}

		// Convert bet sum to points (assuming bets are in ETH: 1 ETH = 10^9 points)
		points := int64(bet.Sum * 1e9) // 1 ETH = 10^9 points
		totalPoints += points
	}
	return totalPoints, nil
}

// parsePrizeValueToPoints parses prize value string and converts to points
// Prize values are now stored as numeric strings (points) from prize_values table
// Returns points as int64
func (s *RatingService) parsePrizeValueToPoints(prizeValue string) (int64, error) {
	// Remove whitespace
	prizeValue = strings.TrimSpace(prizeValue)
	if prizeValue == "" {
		return 0, fmt.Errorf("prize value is empty")
	}

	// Parse as integer (prize values are now stored as points directly)
	value, err := strconv.ParseInt(prizeValue, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse prize value as integer: %w", err)
	}

	return value, nil
}

func (s *RatingService) GetGlobalRating(ctx context.Context, limit, offset int) ([]domain.GlobalRatingEntry, error) {
	if s.repo == nil {
		return nil, errors.New("rating repository is not configured")
	}

	if limit <= 0 {
		limit = 50 // Default limit
	}
	if limit > 1000 {
		limit = 1000 // Max limit to prevent abuse
	}
	if offset < 0 {
		offset = 0
	}

	return s.repo.GetGlobalRating(ctx, limit, offset)
}

func (s *RatingService) GetFriendsRatings(ctx context.Context, userUUID string, limit, offset int) ([]domain.FriendRatingEntry, error) {
	if userUUID == "" {
		return nil, errors.New("user uuid is required")
	}

	if s.repo == nil {
		return nil, errors.New("rating repository is not configured")
	}

	if limit <= 0 {
		limit = 50 // Default limit
	}
	if limit > 1000 {
		limit = 1000 // Max limit to prevent abuse
	}
	if offset < 0 {
		offset = 0
	}

	return s.repo.GetFriendsRatings(ctx, userUUID, limit, offset)
}
