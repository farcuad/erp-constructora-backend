package suppliers

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

// CreateSupplier valida y crea un nuevo proveedor
func (s *Service) CreateSupplier(ctx context.Context, supplier *Supplier) error {
	if supplier.Name == "" || supplier.NIT == "" {
		return errors.New("el nombre y el NIT del proveedor son campos obligatorios")
	}
	return s.repo.CreateSupplier(ctx, supplier)
}

// ListSuppliers extrae los proveedores vinculados a la constructora actual
func (s *Service) ListSuppliers(ctx context.Context, companyID string) ([]Supplier, error) {
	return s.repo.GetSuppliersByCompany(ctx, companyID)
}
