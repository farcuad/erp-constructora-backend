package contractors

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

func (s *Service) CreateContractor(ctx context.Context, c *Contractor) error {
	if c.Name == "" || c.NIT == "" {
		return errors.New("la razón social y la identificación fiscal (NIT) son requeridas")
	}
	return s.repo.CreateContractor(ctx, c)
}

func (s *Service) CreateContract(ctx context.Context, cc *ContractorContract) error {
	if cc.ContractorID == "" || cc.ProjectID == "" || cc.TotalAmount <= 0 || cc.StartDate == "" {
		return errors.New("contratista, proyecto, monto base (mayor a cero) y fecha de inicio son mandatorios")
	}
	return s.repo.CreateContract(ctx, cc)
}

func (s *Service) AddPayment(ctx context.Context, p *ContractorPayment) error {
	if p.ContractID == "" || p.Amount <= 0 || p.PaymentDate == "" {
		return errors.New("id del contrato, fecha y monto del desembolso válidos son requeridos")
	}
	return s.repo.RegisterPayment(ctx, p) // Llama al repositorio transaccional
}

func (s *Service) GetAllContractPayment(ctx context.Context) ([]ContractorPayment, error) {
	return s.repo.GetContractPayment(ctx)
}

func (s *Service) ListContractsByProject(ctx context.Context, projectID string) ([]ContractorContract, error) {
	return s.repo.GetContractsByProject(ctx, projectID)
}

func (s *Service) GetALlContract(ctx context.Context, companyID string) ([]Contractor, error) {
	return s.repo.GetContracts(ctx, companyID)
}

func (s *Service) UpdateContractor(ctx context.Context, c *Contractor) error {
	if c.Name == "" || c.NIT == "" {
		return errors.New("la razón social y la identificación fiscal (NIT) son requeridas")
	}
	return s.repo.UpdateContractor(ctx, c)
}

func (s *Service) DeleteContractor(ctx context.Context, companyID, id string) error {
	return s.repo.DeleteContractor(ctx, companyID, id)
}

func (s *Service) UpdateContract(ctx context.Context, cc *ContractorContract) error {
	if cc.Title == "" || cc.TotalAmount <= 0 || cc.StartDate == "" {
		return errors.New("título, monto (mayor a cero) y fecha de inicio son mandatorios")
	}
	return s.repo.UpdateContract(ctx, cc)
}

func (s *Service) DeleteContract(ctx context.Context, companyID, id string) error {
	return s.repo.DeleteContract(ctx, companyID, id)
}
