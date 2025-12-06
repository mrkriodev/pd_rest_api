package data

import "pdrest/internal/domain"

type ClientRepository interface {
	GetStatus(id int) (*domain.ClientStatus, error)
}

type InMemoryClientRepository struct {
	storage map[int]string
}

func NewInMemoryClientRepository() *InMemoryClientRepository {
	return &InMemoryClientRepository{
		storage: map[int]string{
			1: "active",
			2: "pending",
			3: "blocked",
		},
	}
}

func (r *InMemoryClientRepository) GetStatus(id int) (*domain.ClientStatus, error) {
	if status, ok := r.storage[id]; ok {
		return &domain.ClientStatus{ID: id, Status: status}, nil
	}
	return nil, nil
}
