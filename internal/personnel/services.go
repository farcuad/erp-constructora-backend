package personnel

import (
	"context"
	"errors"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreatePosition(ctx context.Context, p *Position) error {
	if p.Name == "" {
		return errors.New("el nombre del cargo es requerido")
	}
	return s.repo.CreatePosition(ctx, p)
}

func (s *Service) RegisterEmployee(ctx context.Context, e *Employee) error {
	if e.FirstName == "" || e.LastName == "" || e.DNI == "" {
		return errors.New("nombre, apellido e identificación (DNI) son mandatorios")
	}
	return s.repo.CreateEmployee(ctx, e)
}

func (s *Service) ListEmployees(ctx context.Context, companyID string) ([]Employee, error) {
	return s.repo.GetEmployeesByCompany(ctx, companyID)
}

func (s *Service) AddContract(ctx context.Context, c *Contract) error {
	if c.EmployeeID == "" || c.ContractType == "" || c.Salary <= 0 || c.StartDate == "" {
		return errors.New("los datos básicos del contrato y un salario mayor a cero son obligatorios")
	}
	return s.repo.CreateContract(ctx, c)
}
