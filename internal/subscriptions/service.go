package subscriptions

import (
	"context"
	"errors"
	"time"

	"erp-constructora/internal/middlewares"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetMySubscription(ctx context.Context, companyID string) (*CompanySubscription, error) {
	if companyID == "" {
		return nil, errors.New("identificador de empresa inválido")
	}

	sub, err := s.repo.GetByCompany(ctx, companyID)
	if err != nil {
		return nil, errors.New("no se encontró una suscripción activa para esta empresa")
	}
	return sub, nil
}

func (s *Service) ActivateSubscription(ctx context.Context, companyID string, req *CreateSubscriptionRequest) (*CompanySubscription, error) {
	if req.MaxProjects <= 0 {
		req.MaxProjects = 1
	}
	if req.MaxUsers <= 0 {
		req.MaxUsers = 3
	}
	if req.BillingCycle == "" {
		req.BillingCycle = "monthly"
	}
	if req.Status == "" {
		req.Status = "trial"
	}

	return s.repo.Create(ctx, companyID, req)
}

func (s *Service) UpdateSubscription(ctx context.Context, id, companyID string, req *UpdateSubscriptionRequest) (*CompanySubscription, error) {
	existing, err := s.repo.GetByCompany(ctx, companyID)
	if err != nil {
		return nil, errors.New("suscripción no encontrada")
	}

	if req.MaxProjects != nil && *req.MaxProjects < 0 {
		return nil, errors.New("max_projects no puede ser negativo")
	}
	if req.MaxUsers != nil && *req.MaxUsers < 0 {
		return nil, errors.New("max_users no puede ser negativo")
	}

	return s.repo.Update(ctx, existing.ID, companyID, req)
}

func (s *Service) GetSubscriptionInfo(ctx context.Context, companyID string) (*middlewares.SubscriptionInfo, error) {
	sub, err := s.repo.GetByCompany(ctx, companyID)
	if err != nil {
		return nil, err
	}

	return &middlewares.SubscriptionInfo{
		ID:          sub.ID,
		CompanyID:   sub.CompanyID,
		Status:      sub.Status,
		MaxProjects: sub.MaxProjects,
		MaxUsers:    sub.MaxUsers,
	}, nil
}

func (s *Service) IsSubscriptionActive(ctx context.Context, companyID string) (bool, error) {
	sub, err := s.repo.GetByCompany(ctx, companyID)
	if err != nil {
		return false, nil
	}

	if sub.Status == "cancelled" || sub.Status == "expired" {
		return false, nil
	}

	if sub.EndDate != nil && sub.EndDate.Before(time.Now()) {
		return false, nil
	}

	return true, nil
}

func (s *Service) CanCreateProject(ctx context.Context, companyID string) (bool, error) {
	sub, err := s.repo.GetByCompany(ctx, companyID)
	if err != nil {
		return false, err
	}

	if sub.MaxProjects == -1 {
		return true, nil
	}

	current, err := s.repo.CountActiveProjects(ctx, companyID)
	if err != nil {
		return false, err
	}

	return current < sub.MaxProjects, nil
}

func (s *Service) CanCreateUser(ctx context.Context, companyID string) (bool, error) {
	sub, err := s.repo.GetByCompany(ctx, companyID)
	if err != nil {
		return false, err
	}

	if sub.MaxUsers == -1 {
		return true, nil
	}

	current, err := s.repo.CountActiveUsers(ctx, companyID)
	if err != nil {
		return false, err
	}

	return current < sub.MaxUsers, nil
}
