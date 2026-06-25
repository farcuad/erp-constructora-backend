package purchase

import (
	"context"
	"errors"
)

type Service interface {
	GeneratePurchaseOrder(ctx context.Context, companyID string, userID string, req *CreatePurchaseOrderRequest) (*PurchaseOrder, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) GeneratePurchaseOrder(ctx context.Context, companyID string, userID string, req *CreatePurchaseOrderRequest) (*PurchaseOrder, error) {
	if req.ProjectID == "" || req.SupplierID == "" {
		return nil, errors.New("el proyecto y el proveedor son requeridos para emitir una orden de compra")
	}
	if len(req.Items) == 0 {
		return nil, errors.New("la orden de compra debe contener por lo menos un ítem")
	}
	return s.repo.CreateOrder(ctx, companyID, userID, req)
}
