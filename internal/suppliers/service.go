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

func (s *Service) UpdateSupplier(ctx context.Context, id, companyID string, req *UpdateSupplierRequest) (*Supplier, error) {
	sup, err := s.repo.GetSupplierByID(ctx, id, companyID)
	if err != nil {
		return nil, err
	}
	if req.Name != nil {
		sup.Name = *req.Name
	}
	if req.NIT != nil {
		sup.NIT = *req.NIT
	}
	if req.Address != nil {
		sup.Address = *req.Address
	}
	if req.Phone != nil {
		sup.Phone = *req.Phone
	}
	if req.Email != nil {
		sup.Email = *req.Email
	}
	if err := s.repo.UpdateSupplier(ctx, sup); err != nil {
		return nil, err
	}
	return sup, nil
}

func (s *Service) DeleteSupplier(ctx context.Context, id, companyID string) error {
	return s.repo.DeleteSupplier(ctx, id, companyID)
}
