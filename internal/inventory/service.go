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
