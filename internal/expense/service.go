package expense

import (
	"context"
	"errors"
)

type Service interface {
	RegisterExpense(ctx context.Context, companyID string, userID string, req *CreateExpenseRequest) (*Expense, error)
	GetProjectExpenses(ctx context.Context, companyID string, projectID string) ([]Expense, error)
	UpdateExpense(ctx context.Context, companyID string, id string, req *UpdateExpenseRequest) (*Expense, error)
	DeleteExpense(ctx context.Context, companyID string, id string) error
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

func (s *service) UpdateExpense(ctx context.Context, companyID string, id string, req *UpdateExpenseRequest) (*Expense, error) {
	if req.Title != nil && *req.Title == "" {
		return nil, errors.New("el título no puede estar vacío")
	}
	if req.Amount != nil && *req.Amount <= 0 {
		return nil, errors.New("el monto debe ser mayor a cero")
	}
	if req.CategoryID != nil && *req.CategoryID <= 0 {
		return nil, errors.New("debe seleccionar una categoría de gasto válida")
	}
	err := s.repo.Update(ctx, companyID, id, req)
	if err != nil {
		return nil, err
	}
	return s.repo.GetByID(ctx, companyID, id)
}

func (s *service) DeleteExpense(ctx context.Context, companyID string, id string) error {
	return s.repo.Delete(ctx, companyID, id)
}

func (s *service) GetProjectExpenses(ctx context.Context, companyID string, projectID string) ([]Expense, error) {
	if companyID == "" || projectID == "" {
		return nil, errors.New("identificadores de empresa o proyecto inválidos")
	}
	return s.repo.GetByProject(ctx, companyID, projectID)
}
