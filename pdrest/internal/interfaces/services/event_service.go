package services

import (
	"context"
	"errors"
	"pdrest/internal/data"
	"pdrest/internal/domain"
	"strconv"
	"time"
)

type EventService struct {
	repo           data.EventRepository
	prizeRepo      data.PrizeRepository
	prizeValueRepo data.PrizeValueRepository
}

func NewEventService(r data.EventRepository, prizeRepo data.PrizeRepository, prizeValueRepo data.PrizeValueRepository) *EventService {
	return &EventService{
		repo:           r,
		prizeRepo:      prizeRepo,
		prizeValueRepo: prizeValueRepo,
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

	hasPrise, _, _, err := s.repo.GetUserEventPrizeStatus(ctx, userUUID, eventID)
	if err != nil {
		return "", err
	}
	if hasPrise != nil {
		return "already_defined", nil
	}

	switch eventID {
	case "example":
		// TODO: implement event-specific checks and set hasPrise/prizeValueID
		hasPrise := false
		var prizeValueID *int

		updated, err := s.repo.UpdateUserEventPrizeStatusIfUnknown(ctx, userUUID, eventID, &hasPrise, prizeValueID)
		if err != nil {
			return "", err
		}
		if !updated {
			return "already_defined", nil
		}
		return "updated", nil
	default:
		return "", errors.New("unsupported event id")
	}
}

func (s *EventService) TakeEventPrize(ctx context.Context, userUUID string, eventID string) (*domain.Prize, error) {
	if userUUID == "" {
		return nil, errors.New("user uuid is required")
	}
	if eventID == "" {
		return nil, errors.New("event_id is required")
	}
	if s.repo == nil || s.prizeRepo == nil || s.prizeValueRepo == nil {
		return nil, errors.New("event service dependencies are not configured")
	}

	_, prizeValueID, prizeTaken, err := s.repo.GetUserEventPrizeStatus(ctx, userUUID, eventID)
	if err != nil {
		return nil, err
	}
	if prizeValueID == nil {
		return nil, errors.New("prize_value_id is not set for this event")
	}
	if prizeTaken != nil && *prizeTaken {
		return nil, errors.New("prize already taken")
	}

	prizeValue, err := s.prizeValueRepo.GetPrizeValueByID(ctx, *prizeValueID)
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
		return nil, err
	}

	updated, err := s.repo.UpdateUserEventPrizeTakenStatusIfNotTaken(ctx, userUUID, eventID, true)
	if err != nil {
		return nil, err
	}
	if !updated {
		return nil, errors.New("prize already taken")
	}

	return prize, nil
}
