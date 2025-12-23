package services

import (
	"context"
	"fmt"
	"log"
	"pdrest/internal/data"
	"sync"
	"time"
)

// BetScheduler manages async timers for bet closing
// It schedules bet closing tasks that fetch prices from Binance after the timeframe expires
type BetScheduler struct {
	repo          data.BetRepository
	priceProvider *PriceProvider
	timers        map[int]*timerInfo
	mu            sync.RWMutex
	ctx           context.Context
	cancel        context.CancelFunc
	wg            sync.WaitGroup
}

type timerInfo struct {
	betID      int
	pair       string
	closeTime  time.Time
	cancelFunc context.CancelFunc
}

// NewBetScheduler creates a new bet scheduler
func NewBetScheduler(repo data.BetRepository, priceProvider *PriceProvider) *BetScheduler {
	ctx, cancel := context.WithCancel(context.Background())
	return &BetScheduler{
		repo:          repo,
		priceProvider: priceProvider,
		timers:        make(map[int]*timerInfo),
		ctx:           ctx,
		cancel:        cancel,
	}
}

// ScheduleBetClosing schedules a bet to be closed after the specified timeframe
// It fetches the current price from Binance when the bet is opened,
// then schedules another fetch after the timeframe expires
func (s *BetScheduler) ScheduleBetClosing(betID int, pair string, openTime time.Time, timeframe int) error {
	if timeframe <= 0 {
		return fmt.Errorf("timeframe must be greater than 0")
	}

	// Calculate when the bet should be closed
	closeTime := openTime.Add(time.Duration(timeframe) * time.Second)
	now := time.Now()

	// If the close time is in the past, close immediately
	if closeTime.Before(now) || closeTime.Equal(now) {
		return s.closeBetImmediately(betID, pair)
	}

	// Create a context for this specific timer
	timerCtx, cancelFunc := context.WithCancel(s.ctx)

	// Store timer info
	s.mu.Lock()
	s.timers[betID] = &timerInfo{
		betID:      betID,
		pair:       pair,
		closeTime:  closeTime,
		cancelFunc: cancelFunc,
	}
	s.mu.Unlock()

	// Calculate duration until close time
	duration := closeTime.Sub(now)

	// Start async goroutine to handle bet closing
	s.wg.Add(1)
	go s.scheduleCloseBet(timerCtx, betID, pair, duration)

	log.Printf("Scheduled bet %d to close at %s (in %v)", betID, closeTime.Format(time.RFC3339), duration)
	return nil
}

// scheduleCloseBet waits for the duration and then closes the bet
func (s *BetScheduler) scheduleCloseBet(ctx context.Context, betID int, pair string, duration time.Duration) {
	defer s.wg.Done()

	// Wait for the duration or context cancellation
	timer := time.NewTimer(duration)
	defer timer.Stop()

	select {
	case <-timer.C:
		// Timeframe expired, close the bet
		if err := s.closeBet(betID, pair); err != nil {
			log.Printf("Error closing bet %d: %v", betID, err)
		}
	case <-ctx.Done():
		// Timer was cancelled
		log.Printf("Bet %d closing timer cancelled", betID)
		return
	}

	// Remove timer from map
	s.mu.Lock()
	delete(s.timers, betID)
	s.mu.Unlock()
}

// closeBet fetches the current price from Binance and updates the bet
func (s *BetScheduler) closeBet(betID int, pair string) error {
	log.Printf("Closing bet %d for pair %s", betID, pair)

	// Fetch current price from Binance
	closePrice, err := s.priceProvider.GetPrice(pair)
	if err != nil {
		return fmt.Errorf("failed to fetch close price for bet %d: %w", betID, err)
	}

	// Update bet with close price
	closeTime := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.repo.UpdateBetClosePrice(ctx, betID, closePrice, closeTime); err != nil {
		return fmt.Errorf("failed to update bet %d close price: %w", betID, err)
	}

	log.Printf("Successfully closed bet %d with price %.8f at %s", betID, closePrice, closeTime.Format(time.RFC3339))
	return nil
}

// closeBetImmediately closes a bet that should have been closed already
func (s *BetScheduler) closeBetImmediately(betID int, pair string) error {
	log.Printf("Closing bet %d immediately (timeframe already expired)", betID)
	return s.closeBet(betID, pair)
}

// CancelBetClosing cancels a scheduled bet closing
func (s *BetScheduler) CancelBetClosing(betID int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if timer, exists := s.timers[betID]; exists {
		timer.cancelFunc()
		delete(s.timers, betID)
		log.Printf("Cancelled bet %d closing timer", betID)
	}
}

// Shutdown gracefully shuts down the scheduler, waiting for all active timers
func (s *BetScheduler) Shutdown() {
	log.Println("Shutting down bet scheduler...")
	s.cancel()
	s.wg.Wait()
	log.Println("Bet scheduler shut down complete")
}

// GetActiveBetsCount returns the number of active bet timers
func (s *BetScheduler) GetActiveBetsCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.timers)
}

