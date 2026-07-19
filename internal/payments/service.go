package payments

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

func (s *Service) SaveInvoice(ctx context.Context, inv *Invoice) error {
	if inv.ProjectID == "" || inv.InvoiceNumber == "" || inv.Type == "" {
		return errors.New("los campos project_id, invoice_number y type son obligatorios")
	}
	if inv.TotalAmount <= 0 {
		return errors.New("el monto total de la factura debe ser mayor a cero")
	}
	inv.Status = "PENDING"
	return s.repo.CreateInvoice(ctx, inv)
}

func (s *Service) ProcessPayment(ctx context.Context, p *Payment) error {
	if p.InvoiceID == "" || p.Amount <= 0 {
		return errors.New("el id de factura y un monto válido son obligatorios")
	}
	return s.repo.RegisterPayment(ctx, p)
}

func (s *Service) UpdateInvoice(ctx context.Context, companyID, id string, req UpdateInvoiceRequest) error {
	if companyID == "" || id == "" {
		return errors.New("el id de la empresa y de la factura son requeridos")
	}
	return s.repo.UpdateInvoice(ctx, companyID, id, req)
}

func (s *Service) DeleteInvoice(ctx context.Context, companyID, id string) error {
	if companyID == "" || id == "" {
		return errors.New("el id de la empresa y de la factura son requeridos")
	}
	return s.repo.DeleteInvoice(ctx, companyID, id)
}

func (s *Service) CancelInvoice(ctx context.Context, companyID, id string) error {
	if companyID == "" || id == "" {
		return errors.New("el id de la empresa y de la factura son requeridos")
	}
	return s.repo.CancelInvoice(ctx, companyID, id)
}

func (s *Service) GetProjectInvoices(ctx context.Context, companyID, projectID string) ([]Invoice, error) {
	if companyID == "" || projectID == "" {
		return nil, errors.New("empresa y proyecto son requeridos")
	}
	return s.repo.GetByProject(ctx, companyID, projectID)
}

func (s *Service) GetInvoiceByID(ctx context.Context, companyID, id string) (*Invoice, error) {
	if companyID == "" || id == "" {
		return nil, errors.New("empresa e id de factura son requeridos")
	}
	return s.repo.GetByID(ctx, companyID, id)
}

func (s *Service) GetPayments(ctx context.Context, companyID, invoiceID string) ([]Payment, error) {
	if companyID == "" || invoiceID == "" {
		return nil, errors.New("empresa e id de factura son requeridos")
	}
	return s.repo.GetPaymentsByInvoice(ctx, companyID, invoiceID)
}
