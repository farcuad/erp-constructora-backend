package purchase

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

// CreatePurchaseOrder calcula los totales y guarda la orden
func (s *Service) CreatePurchaseOrder(ctx context.Context, po *PurchaseOrder) error {
	if po.ProjectID == "" || po.SupplierID == "" {
		return errors.New("el proyecto y el proveedor son campos obligatorios")
	}
	if len(po.Items) == 0 {
		return errors.New("la orden de compra debe contener al menos un ítem")
	}

	// Forzar estado por defecto y recalcular la suma total del lado del servidor por seguridad
	po.Status = "Pending"
	var total float64 = 0
	for _, item := range po.Items {
		if item.Quantity <= 0 || item.UnitPrice <= 0 {
			return errors.New("la cantidad y el precio unitario deben ser mayores a cero")
		}
		total += item.Quantity * item.UnitPrice
	}
	po.TotalAmount = total

	return s.repo.CreatePurchaseOrder(ctx, po)
}

// ListOrdersByProject lista las órdenes correspondientes a una obra
func (s *Service) ListOrdersByProject(ctx context.Context, projectID string) ([]PurchaseOrder, error) {
	return s.repo.GetOrdersByProject(ctx, projectID)
}
