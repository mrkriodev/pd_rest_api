package services

import (
	"context"
	"errors"
	"fmt"
	"pdrest/internal/data"
	"pdrest/internal/domain"
	"strconv"
	"strings"
	"time"
)

type AchievementService struct {
	repo           data.AchievementRepository
	prizeRepo      data.PrizeRepository
	prizeValueRepo data.PrizeValueRepository
	ratingRepo     data.RatingRepository
	betRepo        data.BetRepository
}

func NewAchievementService(r data.AchievementRepository, prizeRepo data.PrizeRepository, prizeValueRepo data.PrizeValueRepository, ratingRepo data.RatingRepository, betRepo data.BetRepository) *AchievementService {
	return &AchievementService{
		repo:           r,
		prizeRepo:      prizeRepo,
		prizeValueRepo: prizeValueRepo,
		ratingRepo:     ratingRepo,
		betRepo:        betRepo,
	}
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

	filtered := achievements[:0]
	for _, achievement := range achievements {
		if !hasTag(achievement.Tags, "event") {
			filtered = append(filtered, achievement)
		}
	}

	return &domain.UserAchievementsResponse{
		Achievements: filtered,
	}, nil
}

func hasTag(tags string, target string) bool {
	if tags == "" || target == "" {
		return false
	}
	for _, tag := range strings.Split(tags, ",") {
		if strings.TrimSpace(tag) == target {
			return true
		}
	}
	return false
}

func (s *AchievementService) UpdateAchievementStatus(ctx context.Context, userUUID string, achievementID string) (string, error) {
	if userUUID == "" {
		return "", errors.New("user uuid is required")
	}
	if achievementID == "" {
		return "", errors.New("achievement_id is required")
	}
	if s.repo == nil || s.betRepo == nil {
		return "", errors.New("achievement service dependencies are not configured")
	}

	switch achievementID {
	case "first_bet_success":
		hasWin, err := s.betRepo.HasWinningBet(ctx, userUUID)
		if err != nil {
			return "", err
		}
		if !hasWin {
			return "not_completed", nil
		}

		achievement, err := s.repo.GetAchievementByID(ctx, achievementID)
		if err != nil {
			return "", err
		}
		needSteps := achievement.Steps
		if needSteps <= 0 {
			needSteps = 1
		}

		created, err := s.repo.AddUserAchievement(ctx, userUUID, achievementID, needSteps, needSteps)
		if err != nil {
			return "", err
		}
		if !created {
			return "already_exists", nil
		}
		return "created", nil
	default:
		return "", errors.New("unsupported achievement id")
	}
}

func (s *AchievementService) ClaimAchievement(ctx context.Context, userUUID string, achievementID string) (*domain.Prize, error) {
	if userUUID == "" {
		return nil, errors.New("user uuid is required")
	}
	if achievementID == "" {
		return nil, errors.New("achievement_id is required")
	}
	if s.repo == nil || s.prizeRepo == nil || s.prizeValueRepo == nil || s.ratingRepo == nil {
		return nil, errors.New("achievement service dependencies are not configured")
	}

	achievement, err := s.repo.GetAchievementByID(ctx, achievementID)
	if err != nil {
		return nil, err
	}
	if achievement == nil {
		return nil, errors.New("achievement not found")
	}

	status, err := s.repo.GetUserAchievementStatus(ctx, userUUID, achievementID)
	if err != nil {
		return nil, err
	}
	if status == nil {
		return nil, errors.New("user achievement not found")
	}
	if status.ClaimedStatus {
		return nil, errors.New("achievement already claimed")
	}

	needSteps := status.NeedSteps
	if needSteps <= 0 {
		needSteps = achievement.Steps
		if needSteps > 0 {
			_ = s.repo.UpdateUserAchievementNeedSteps(ctx, userUUID, achievementID, needSteps)
		}
	}
	if status.StepsGot < needSteps {
		return nil, errors.New("achievement is not completed yet")
	}

	if achievement.PrizeID == nil {
		return nil, errors.New("achievement has no prize")
	}

	prizeValue, err := s.prizeValueRepo.GetPrizeValueByID(ctx, *achievement.PrizeID)
	if err != nil {
		return nil, err
	}
	if prizeValue == nil {
		return nil, errors.New("prize value not found")
	}

	prizeValueStr := prizeValue.Label
	if prizeValueStr == "" {
		prizeValueStr = strconv.FormatInt(prizeValue.Value, 10)
	}

	now := time.Now().UnixMilli()
	eventID := prizeValue.EventID
	prize := &domain.Prize{
		EventID:      &eventID,
		UserID:       &userUUID,
		PrizeValueID: &prizeValue.ID,
		PrizeValue:   prizeValueStr,
		PrizeType:    domain.PrizeTypeEventReward,
		AwardedAt:    now,
		CreatedAt:    now,
	}

	if err := s.prizeRepo.CreatePrize(ctx, prize); err != nil {
		return nil, fmt.Errorf("failed to create prize record: %w", err)
	}

	points := prizeValue.Value
	description := fmt.Sprintf("Achievement %s: %d points", achievement.ID, points)
	if err := s.ratingRepo.AddPoints(ctx, userUUID, points, &prize.ID, nil, description); err != nil {
		return nil, fmt.Errorf("failed to add achievement points: %w", err)
	}

	if err := s.repo.UpdateUserAchievementClaimStatus(ctx, userUUID, achievementID, true); err != nil {
		return nil, fmt.Errorf("failed to update achievement claim: %w", err)
	}

	return prize, nil
}
