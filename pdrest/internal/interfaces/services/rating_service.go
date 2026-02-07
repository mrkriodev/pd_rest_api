package services

import (
	"context"
	"errors"
	"pdrest/internal/data"
	"pdrest/internal/domain"
)

type RatingService struct {
	repo data.RatingRepository
}

func NewRatingService(r data.RatingRepository) *RatingService {
	return &RatingService{
		repo: r,
	}
}

func (s *RatingService) GetUserAssets(ctx context.Context, userUUID string) (*domain.UserAssets, error) {
	if userUUID == "" {
		return nil, errors.New("user uuid is required")
	}

	if s.repo == nil {
		return nil, errors.New("rating repository is not configured")
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
