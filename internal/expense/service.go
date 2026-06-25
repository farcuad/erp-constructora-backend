package expense

import (
	"context"
	"errors"
)

type Service interface {
	RegisterExpense(ctx context.Context, companyID string, userID string, req *CreateExpenseRequest) (*Expense, error)
	GetProjectExpenses(ctx context.Context, companyID string, projectID string) ([]Expense, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) RegisterExpense(ctx context.Context, companyID string, userID string, req *CreateExpenseRequest) (*Expense, error) {
	if req.ProjectID == "" || req.Title == "" || req.Amount <= 0 {
		return nil, errors.New("el proyecto, el título y un monto mayor a cero son obligatorios")
	}
	if req.CategoryID <= 0 {
		return nil, errors.New("debe seleccionar una categoría de gasto válida")
	}
	return s.repo.Create(ctx, companyID, userID, req)
}

func (s *service) GetProjectExpenses(ctx context.Context, companyID string, projectID string) ([]Expense, error) {
	if companyID == "" || projectID == "" {
		return nil, errors.New("identificadores de empresa o proyecto inválidos")
	}
	return s.repo.GetByProject(ctx, companyID, projectID)
}
