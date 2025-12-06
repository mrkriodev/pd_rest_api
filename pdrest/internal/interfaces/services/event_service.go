package services

import (
	"context"
	"pdrest/internal/data"
	"pdrest/internal/domain"
)

type EventService struct {
	repo data.EventRepository
}

func NewEventService(r data.EventRepository) *EventService {
	return &EventService{repo: r}
}

func (s *EventService) GetAvailableEvents(ctx context.Context) ([]domain.Event, error) {
	return s.repo.GetAllEvents(ctx)
}
