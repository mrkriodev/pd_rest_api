package services

import (
	"context"
	"errors"
	"pdrest/internal/data"
	"pdrest/internal/domain"
)

type AchievementService struct {
	repo data.AchievementRepository
}

func NewAchievementService(r data.AchievementRepository) *AchievementService {
	return &AchievementService{repo: r}
}

func (s *AchievementService) GetAvailableAchievements(ctx context.Context) (*domain.AchievementsResponse, error) {
	if s.repo == nil {
		return nil, errors.New("achievement repository is not configured")
	}

	achievements, err := s.repo.GetAllAchievements(ctx)
	if err != nil {
		return nil, err
	}

	return &domain.AchievementsResponse{
		Achievements: achievements,
	}, nil
}

func (s *AchievementService) GetUserAchievements(ctx context.Context, userUUID string) (*domain.UserAchievementsResponse, error) {
	if userUUID == "" {
		return nil, errors.New("user uuid is required")
	}

	if s.repo == nil {
		return nil, errors.New("achievement repository is not configured")
	}

	achievements, err := s.repo.GetUserAchievements(ctx, userUUID)
	if err != nil {
		return nil, err
	}

	return &domain.UserAchievementsResponse{
		Achievements: achievements,
	}, nil
}
