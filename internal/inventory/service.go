package inventory

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

func (s *Service) CreateWarehouse(ctx context.Context, w *Warehouse) error {
	if w.ProjectID == "" || w.Name == "" {
		return errors.New("el proyecto y el nombre de la bodega son requeridos")
	}
	return s.repo.CreateWarehouse(ctx, w)
}

func (s *Service) CreateMaterial(ctx context.Context, m *Material) error {
	if m.Name == "" || m.Unit == "" {
		return errors.New("el nombre del material y la unidad de medida son requeridos")
	}
	return s.repo.CreateMaterial(ctx, m)
}

func (s *Service) RegisterMovement(ctx context.Context, m *StockMovement) error {
	if m.WarehouseID == "" || m.MaterialID == "" || m.Quantity <= 0 {
		return errors.New("bodega, material y cantidad (mayor a cero) son mandatorios")
	}
	if m.MovementType != "INPUT" && m.MovementType != "OUTPUT" {
		return errors.New("el tipo de movimiento debe ser 'INPUT' o 'OUTPUT'")
	}
	return s.repo.ExecuteStockMovement(ctx, m)
}

func (s *Service) GetCurrentStock(ctx context.Context, warehouseID string) ([]MaterialStock, error) {
	return s.repo.GetStockByWarehouse(ctx, warehouseID)
}

func (s *Service) GetMaterials(ctx context.Context, companyID string) (*Material, error) {
	return s.repo.GetMaterials(ctx, companyID)
}

func (s *Service) UpdateMaterial(ctx context.Context, id, companyID string, req *UpdateMaterialRequest) (*Material, error) {
	m, err := s.repo.GetMaterialByID(ctx, id, companyID)
	if err != nil {
		return nil, err
	}
	if req.Name != nil {
		m.Name = *req.Name
	}
	if req.Code != nil {
		m.Code = *req.Code
	}
	if req.Unit != nil {
		m.Unit = *req.Unit
	}
	if req.CategoryID != nil {
		m.CategoryID = *req.CategoryID
	}
	if err := s.repo.UpdateMaterial(ctx, m); err != nil {
		return nil, err
	}
	return m, nil
}

func (s *Service) DeleteMaterial(ctx context.Context, id, companyID string) error {
	return s.repo.DeleteMaterial(ctx, id, companyID)
}

func (s *Service) UpdateWarehouse(ctx context.Context, id, companyID string, req *UpdateWarehouseRequest) (*Warehouse, error) {
	w, err := s.repo.GetWarehouseByID(ctx, id, companyID)
	if err != nil {
		return nil, err
	}
	if req.Name != nil {
		w.Name = *req.Name
	}
	if req.Location != nil {
		w.Location = *req.Location
	}
	if err := s.repo.UpdateWarehouse(ctx, w); err != nil {
		return nil, err
	}
	return w, nil
}

func (s *Service) DeleteWarehouse(ctx context.Context, id, companyID string) error {
	return s.repo.DeleteWarehouse(ctx, id, companyID)
}
