package purchase

import (
	"context"
	"encoding/json"
	"erp-constructora/internal/middlewares"
	"erp-constructora/internal/notifications"
	"erp-constructora/internal/utils"
	"fmt"
	"log"
	"net/http"
)

func strPtr(s string) *string {
	return &s
}

type Handler struct {
	service  *Service
	notifier notifications.Notifier
}

func NewHandler(service *Service, notifier notifications.Notifier) *Handler {
	return &Handler{service: service, notifier: notifier}
}

// notify emite una notificación a toda la empresa (el actor queda excluido por NotifyFromContext)
func (h *Handler) notify(ctx context.Context, req notifications.CreateNotificationRequest) {
	if h.notifier == nil {
		return
	}
	if err := h.notifier.NotifyFromContext(ctx, req); err != nil {
		log.Printf("[NOTIFY ERROR] purchase: %v", err)
	}
}

// --- CONTROLADORES DE ÓRDENES DE COMPRA ---

func (h *Handler) CreatePurchaseOrder(w http.ResponseWriter, r *http.Request) {
	// 1. Extraer tanto el company_id como el user_id usando tus helpers nativos
	companyID, okCompany := middlewares.GetCompanyIDFromContext(r.Context())
	userID, okUser := middlewares.GetUserIDFromContext(r.Context())

	if !okCompany || !okUser {
		utils.WriteUnauthorized(w)
		return
	}

	var po PurchaseOrder
	if err := json.NewDecoder(r.Body).Decode(&po); err != nil {
		utils.WriteBadRequest(w, "JSON inválido: "+err.Error())
		return
	}

	// Inyectar de forma segura los datos extraídos del JWT de la sesión
	po.CompanyID = companyID
	po.UserID = userID

	if err := h.service.CreatePurchaseOrder(r.Context(), &po); err != nil {
		utils.WriteBadRequest(w, utils.GetPGErrorMessage(err))
		return
	}

	h.notify(r.Context(), notifications.CreateNotificationRequest{
		ProjectID:  &po.ProjectID,
		EntityType: "PURCHASE_ORDER",
		EntityID:   &po.ID,
		Type:       "PURCHASE_ORDER_CREATED",
		Priority:   notifications.PriorityMedium,
		Title:      "Nueva orden de compra",
		Message:    fmt.Sprintf("Se creó una orden de compra por $%.2f.", po.TotalAmount),
		LinkToUI:   strPtr("/dashboard/projects/" + po.ProjectID + "/purchase"),
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(po)
}

func (h *Handler) GetOrdersByProject(w http.ResponseWriter, r *http.Request) {
	// En Go 1.22+, extraemos el parámetro dinámico {project_id} definido en router.go
	projectID := r.PathValue("project_id")
	if projectID == "" {
		utils.WriteBadRequest(w, "El parámetro project_id es obligatorio")
		return
	}

	orders, err := h.service.ListOrdersByProject(r.Context(), projectID)
	if err != nil {
		utils.WriteInternalError(w, utils.GetPGErrorMessage(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(orders)
}

func (h *Handler) UpdatePurchaseOrder(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		utils.WriteBadRequest(w, "El parámetro id es obligatorio")
		return
	}

	companyID, ok := middlewares.GetCompanyIDFromContext(r.Context())
	if !ok {
		utils.WriteUnauthorized(w)
		return
	}

	var req UpdatePurchaseOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteBadRequest(w, "JSON inválido: "+err.Error())
		return
	}

	po, err := h.service.UpdatePurchaseOrder(r.Context(), id, companyID, &req)
	if err != nil {
		utils.WriteBadRequest(w, utils.GetPGErrorMessage(err))
		return
	}

	h.notify(r.Context(), notifications.CreateNotificationRequest{
		EntityType: "PURCHASE_ORDER",
		EntityID:   &id,
		Type:       "PURCHASE_ORDER_UPDATED",
		Priority:   notifications.PriorityMedium,
		Title:      "Orden de compra actualizada",
		Message:    "Se actualizó una orden de compra.",
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(po)
}

func (h *Handler) DeletePurchaseOrder(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		utils.WriteBadRequest(w, "El parámetro id es obligatorio")
		return
	}

	companyID, ok := middlewares.GetCompanyIDFromContext(r.Context())
	if !ok {
		utils.WriteUnauthorized(w)
		return
	}

	if err := h.service.DeletePurchaseOrder(r.Context(), id, companyID); err != nil {
		utils.WriteBadRequest(w, utils.GetPGErrorMessage(err))
		return
	}

	h.notify(r.Context(), notifications.CreateNotificationRequest{
		EntityType: "PURCHASE_ORDER",
		EntityID:   &id,
		Type:       "PURCHASE_ORDER_DELETED",
		Priority:   notifications.PriorityMedium,
		Title:      "Orden de compra eliminada",
		Message:    "Se eliminó una orden de compra.",
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "recurso eliminado"})
}
