package clients

import (
	"context"
	"errors"
)

type Service interface {
	CreateClient(ctx context.Context, companyID string, req *CreateClientRequest) (*Client, error)
	GetClients(ctx context.Context, companyID string) ([]Client, error)
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

func (s *service) GetClients(ctx context.Context, companyID string) ([]Client, error) {
	if companyID == "" {
		return nil, errors.New("invalid company id")
	}
	return s.repo.GetByCompany(ctx, companyID)
}
