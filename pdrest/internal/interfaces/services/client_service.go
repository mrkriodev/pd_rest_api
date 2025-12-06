package services

import (
	"errors"
	"pdrest/internal/data"
	"pdrest/internal/domain"
)

type ClientService struct {
	repo data.ClientRepository
}

func NewClientService(r data.ClientRepository) *ClientService {
	return &ClientService{repo: r}
}

func (s *ClientService) GetStatus(id int) (*domain.ClientStatus, error) {
	result, err := s.repo.GetStatus(id)
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, errors.New("client not found")
	}
	return result, nil
}
