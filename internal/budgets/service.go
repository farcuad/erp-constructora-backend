package budgets

import (
	"context"
	"errors"
)

type Service interface {
	CreateInitialBudget(ctx context.Context, companyID string, userID string, req *CreateBudgetWithItemsRequest) (*Budget, error)
	GetBudgetsProjectID(ctx context.Context, companyID string, projectID string) ([]Budget, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) CreateInitialBudget(ctx context.Context, companyID string, userID string, req *CreateBudgetWithItemsRequest) (*Budget, error) {
	if req.ProjectID == "" || req.Title == "" {
		return nil, errors.New("el ID del proyecto y el título son obligatorios")
	}
	if len(req.Items) == 0 {
		return nil, errors.New("el presupuesto debe contener al menos un ítem")
	}
	return s.repo.CreateWithItems(ctx, companyID, userID, req)
}

func (s *service) GetBudgetsProjectID(ctx context.Context, companyID string, projectID string) ([]Budget, error) {
	return s.repo.GetBudgetsProjectID(ctx, companyID, projectID)
}
