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

func (s *Service) GetPositions(ctx context.Context, companyID string) ([]Position, error) {
	return s.repo.GetPosition(ctx, companyID)
}

func (s *Service) UpdatePosition(ctx context.Context, p *Position) error {
	if p.Name == "" {
		return errors.New("el nombre del cargo es requerido")
	}
	return s.repo.UpdatePosition(ctx, p)
}

func (s *Service) DeletePosition(ctx context.Context, companyID, id string) error {
	return s.repo.DeletePosition(ctx, companyID, id)
}

func (s *Service) RegisterEmployee(ctx context.Context, e *Employee) error {
	if e.FirstName == "" || e.LastName == "" || e.DNI == "" {
		return errors.New("nombre, apellido e identificación (DNI) son mandatorios")
	}
	return s.repo.CreateEmployee(ctx, e)
}

func (s *Service) UpdateEmployee(ctx context.Context, e *Employee) error {
	if e.FirstName == "" || e.LastName == "" || e.DNI == "" {
		return errors.New("nombre, apellido e identificación (DNI) son mandatorios")
	}
	return s.repo.UpdateEmployee(ctx, e)
}

func (s *Service) DeleteEmployee(ctx context.Context, companyID, id string) error {
	return s.repo.DeleteEmployee(ctx, companyID, id)
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

func (s *Service) GetAllContract(ctx context.Context, projectID string) ([]Contract, error) {
	return s.repo.GetContract(ctx, projectID)
}

func (s *Service) UpdateContract(ctx context.Context, c *Contract) error {
	if c.ContractType == "" || c.Salary <= 0 || c.StartDate == "" {
		return errors.New("los datos básicos del contrato y un salario mayor a cero son obligatorios")
	}
	return s.repo.UpdateContract(ctx, c)
}

func (s *Service) DeleteContract(ctx context.Context, id string) error {
	return s.repo.DeleteContract(ctx, id)
}
