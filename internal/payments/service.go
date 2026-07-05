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
