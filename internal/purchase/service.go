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

func (s *Service) UpdatePurchaseOrder(ctx context.Context, id, companyID string, req *UpdatePurchaseOrderRequest) (*PurchaseOrder, error) {
	po, err := s.repo.GetPurchaseOrderByID(ctx, id, companyID)
	if err != nil {
		return nil, err
	}
	if req.Status != nil {
		po.Status = *req.Status
	}
	if req.DeliveryDate != nil {
		po.DeliveryDate = *req.DeliveryDate
	}
	if req.Notes != nil {
		po.Notes = *req.Notes
	}
	if err := s.repo.UpdatePurchaseOrder(ctx, po); err != nil {
		return nil, err
	}
	return po, nil
}

func (s *Service) DeletePurchaseOrder(ctx context.Context, id, companyID string) error {
	return s.repo.DeletePurchaseOrder(ctx, id, companyID)
}
