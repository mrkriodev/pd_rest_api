package services

import (
	"context"
	"errors"
	"pdrest/internal/data"
	"pdrest/internal/domain"
	"sort"
	"strconv"
	"strings"
	"time"
)

type EventService struct {
	repo            data.EventRepository
	prizeRepo       data.PrizeRepository
	prizeValueRepo  data.PrizeValueRepository
	achievementRepo data.AchievementRepository
	ratingRepo      data.RatingRepository
}

func NewEventService(r data.EventRepository, prizeRepo data.PrizeRepository, prizeValueRepo data.PrizeValueRepository, achievementRepo data.AchievementRepository, ratingRepo data.RatingRepository) *EventService {
	return &EventService{
		repo:            r,
		prizeRepo:       prizeRepo,
		prizeValueRepo:  prizeValueRepo,
		achievementRepo: achievementRepo,
		ratingRepo:      ratingRepo,
	}
}

func (s *EventService) GetAvailableEvents(ctx context.Context, tag string) ([]domain.Event, error) {
	return s.repo.GetAllEvents(ctx, tag)
}

func (s *EventService) TakePartOnEvent(ctx context.Context, userUUID string, eventID string) (string, error) {
	if userUUID == "" {
		return "", errors.New("user uuid is required")
	}
	if eventID == "" {
		return "", errors.New("event_id is required")
	}

	event, err := s.repo.GetEventByID(ctx, eventID)
	if err != nil {
		return "", err
	}
	if event == nil {
		return "", errors.New("event not found")
	}

	created, err := s.repo.AddUserEvent(ctx, userUUID, eventID, "joined")
	if err != nil {
		return "", err
	}
	if !created {
		return "already_exists", nil
	}

	return "created", nil
}

func (s *EventService) GetUserEvents(ctx context.Context, userUUID string) (*domain.UserEventsResponse, error) {
	if userUUID == "" {
		return nil, errors.New("user uuid is required")
	}

	nowMs := time.Now().UTC().UnixMilli()
	events, err := s.repo.GetUserEventsWithAvailable(ctx, userUUID, "competition", nowMs)
	if err != nil {
		return nil, err
	}

	return &domain.UserEventsResponse{
		Events: events,
	}, nil
}

func (s *EventService) UpdateUserEventPrizeStatus(ctx context.Context, userUUID string, eventID string) (string, error) {
	if userUUID == "" {
		return "", errors.New("user uuid is required")
	}
	if eventID == "" {
		return "", errors.New("event_id is required")
	}
	if s.repo == nil || s.ratingRepo == nil || s.prizeValueRepo == nil {
		return "", errors.New("event service dependencies are not configured")
	}

	event, err := s.repo.GetEventByID(ctx, eventID)
	if err != nil {
		return "", err
	}
	if event == nil {
		return "", errors.New("event not found")
	}
	if time.Now().UTC().Before(event.Deadline) {
		return "", errors.New("event is not finished yet")
	}
	startMs := event.StartTime.UTC().UnixMilli()

	hasPriseStatus, _, _, err := s.repo.GetUserEventPrizeStatus(ctx, userUUID, eventID)
	if err != nil {
		return "", err
	}
	if hasPriseStatus != nil {
		return "already_defined", nil
	}

	leaderboard, err := s.ratingRepo.GetBetPointsLeaderboard(ctx, startMs, event.Deadline.UTC().UnixMilli(), 3)
	if err != nil {
		return "", err
	}

	hasPrise := false
	var prizeValueID *int

	for idx, entry := range leaderboard {
		if entry.UserUUID != userUUID {
			continue
		}
		hasPrise = true

		prizeValue, err := s.prizeValueRepo.GetPrizeValuesByEventID(ctx, eventID)
		if err != nil {
			return "", err
		}
		if len(prizeValue) == 0 {
			return "", errors.New("prize values not found for event")
		}

		// Sort by value descending and pick index (0-based) for place.
		sort.Slice(prizeValue, func(i, j int) bool {
			return prizeValue[i].Value > prizeValue[j].Value
		})
		if idx >= len(prizeValue) {
			break
		}
		prizeValueID = &prizeValue[idx].ID
		break
	}

	updated, err := s.repo.UpdateUserEventPrizeStatusIfUnknown(ctx, userUUID, eventID, &hasPrise, prizeValueID)
	if err != nil {
		return "", err
	}
	if !updated {
		return "already_defined", nil
	}
	return "updated", nil
}

func (s *EventService) TakeEventPrize(ctx context.Context, userUUID string, eventID string) (*domain.Prize, string, error) {
	if userUUID == "" {
		return nil, "", errors.New("user uuid is required")
	}
	if eventID == "" {
		return nil, "", errors.New("event_id is required")
	}
	if s.repo == nil || s.prizeRepo == nil || s.prizeValueRepo == nil || s.achievementRepo == nil {
		return nil, "", errors.New("event service dependencies are not configured")
	}

	_, prizeValueID, prizeTaken, err := s.repo.GetUserEventPrizeStatus(ctx, userUUID, eventID)
	if err != nil {
		return nil, "", err
	}
	if prizeValueID == nil {
		return nil, "", errors.New("prize_value_id is not set for this event")
	}
	if prizeTaken != nil && *prizeTaken {
		return nil, "", errors.New("prize already taken")
	}

	prizeValue, err := s.prizeValueRepo.GetPrizeValueByID(ctx, *prizeValueID)
	if err != nil {
		return nil, "", err
	}
	if prizeValue == nil {
		return nil, "", errors.New("prize value not found")
	}

	prizeValueStr := prizeValue.Label
	if prizeValueStr == "" {
		prizeValueStr = strconv.FormatInt(prizeValue.Value, 10)
	}

	now := time.Now().UTC().UnixMilli()
	prize := &domain.Prize{
		EventID:      &eventID,
		UserID:       &userUUID,
		PrizeValueID: prizeValueID,
		PrizeValue:   prizeValueStr,
		PrizeType:    domain.PrizeTypeEventReward,
		AwardedAt:    now,
		CreatedAt:    now,
	}

	if err := s.prizeRepo.CreatePrize(ctx, prize); err != nil {
		return nil, "", err
	}

	updated, err := s.repo.UpdateUserEventPrizeTakenStatusIfNotTaken(ctx, userUUID, eventID, true)
	if err != nil {
		return nil, "", err
	}
	if !updated {
		return nil, "", errors.New("prize already taken")
	}

	var achievementImageURL string
	if prizeValueID != nil {
		achievement, err := s.achievementRepo.GetAchievementByPrizeID(ctx, *prizeValueID)
		if err != nil {
			return nil, "", err
		}
		if achievement != nil {
			if err := s.achievementRepo.UpsertUserAchievementProgress(ctx, userUUID, achievement.ID, 1, 1, true); err != nil {
				return nil, "", err
			}
			achievementImageURL = achievement.ImageURL
		}
	}

	return prize, achievementImageURL, nil
}

func (s *EventService) GetUserEventProgress(ctx context.Context, userUUID string, eventID string) (*domain.EventProgressResponse, error) {
	if userUUID == "" {
		return nil, errors.New("user uuid is required")
	}
	if eventID == "" {
		return nil, errors.New("event_id is required")
	}
	if s.repo == nil || s.ratingRepo == nil {
		return nil, errors.New("event service dependencies are not configured")
	}

	event, err := s.repo.GetEventByID(ctx, eventID)
	if err != nil {
		return nil, err
	}
	if event == nil {
		return nil, errors.New("event not found")
	}
	if !strings.Contains(strings.ToLower(event.Tags), "competition") {
		return nil, errors.New("event is not a competition")
	}

	startMs := event.StartTime.UTC().UnixMilli()
	endMs := event.Deadline.UTC().UnixMilli()

	nowMs := time.Now().UTC().UnixMilli()
	if nowMs < startMs {
		return nil, errors.New("event is not started yet")
	}

	participating, err := s.repo.HasUserEvent(ctx, userUUID, eventID)
	if err != nil {
		return nil, err
	}
	if !participating {
		return &domain.EventProgressResponse{
			EventID:         eventID,
			Participating:   false,
			CollectedPoints: 0,
		}, nil
	}

	points, err := s.ratingRepo.GetUserBetPointsInRange(ctx, userUUID, startMs, endMs)
	if err != nil {
		return nil, err
	}

	return &domain.EventProgressResponse{
		EventID:         eventID,
		Participating:   true,
		CollectedPoints: points,
	}, nil
}

func (s *EventService) GetBestInEvent(ctx context.Context, eventID string) (*domain.EventLeaderResponse, error) {
	if eventID == "" {
		return nil, errors.New("event_id is required")
	}
	if s.repo == nil || s.ratingRepo == nil || s.prizeValueRepo == nil || s.achievementRepo == nil {
		return nil, errors.New("event service dependencies are not configured")
	}

	event, err := s.repo.GetEventByID(ctx, eventID)
	if err != nil {
		return nil, err
	}
	if event == nil {
		return nil, errors.New("event not found")
	}
	if !strings.Contains(strings.ToLower(event.Tags), "competition") {
		return nil, errors.New("event is not a competition")
	}

	startMs := event.StartTime.UTC().UnixMilli()
	endMs := event.Deadline.UTC().UnixMilli()

	nowMs := time.Now().UTC().UnixMilli()
	if nowMs < startMs || nowMs >= endMs {
		return nil, errors.New("event is not active")
	}

	leaders, err := s.ratingRepo.GetBetPointsLeaderboard(ctx, startMs, endMs, 1)
	if err != nil {
		return nil, err
	}
	if len(leaders) == 0 {
		return &domain.EventLeaderResponse{
			LeaderImage: "",
			Points:      0,
		}, nil
	}

	prizeValues, err := s.prizeValueRepo.GetPrizeValuesByEventID(ctx, eventID)
	if err != nil {
		return nil, err
	}
	if len(prizeValues) == 0 {
		return nil, errors.New("prize values not found for event")
	}
	sort.Slice(prizeValues, func(i, j int) bool {
		return prizeValues[i].Value > prizeValues[j].Value
	})

	var leaderImage string
	achievement, err := s.achievementRepo.GetAchievementByPrizeID(ctx, prizeValues[0].ID)
	if err != nil {
		return nil, err
	}
	if achievement != nil {
		leaderImage = achievement.ImageURL
	}

	return &domain.EventLeaderResponse{
		LeaderImage: leaderImage,
		Points:      leaders[0].NetPoints,
	}, nil
}
