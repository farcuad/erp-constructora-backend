package clients

import (
	"context"
	"errors"
)

type Service interface {
	CreateClient(ctx context.Context, companyID string, req *CreateClientRequest) (*Client, error)
	GetClients(ctx context.Context, companyID string) ([]Client, error)
	UpdateClient(ctx context.Context, companyID string, id string, req *UpdateClientRequest) (*Client, error)
	DeleteClient(ctx context.Context, companyID string, id string) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) CreateClient(ctx context.Context, companyID string, req *CreateClientRequest) (*Client, error) {
	if req.Name == "" || req.NIT == "" {
		return nil, errors.New("el nombre y el NIT son obligatorios")
	}
	return s.repo.Create(ctx, companyID, req)
}

func (s *service) UpdateClient(ctx context.Context, companyID string, id string, req *UpdateClientRequest) (*Client, error) {
	if req.Name != nil && *req.Name == "" {
		return nil, errors.New("el nombre no puede estar vacío")
	}
	if req.NIT != nil && *req.NIT == "" {
		return nil, errors.New("el NIT no puede estar vacío")
	}
	err := s.repo.Update(ctx, companyID, id, req)
	if err != nil {
		return nil, err
	}
	return s.repo.GetByID(ctx, companyID, id)
}

func (s *service) DeleteClient(ctx context.Context, companyID string, id string) error {
	return s.repo.Delete(ctx, companyID, id)
}

func (s *service) GetClients(ctx context.Context, companyID string) ([]Client, error) {
	if companyID == "" {
		return nil, errors.New("invalid company id")
	}
	return s.repo.GetByCompany(ctx, companyID)
}
