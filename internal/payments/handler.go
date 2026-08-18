package payments

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"erp-constructora/internal/middlewares"
	"erp-constructora/internal/notifications"
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
		log.Printf("[NOTIFY ERROR] payments: %v", err)
	}
}

func (h *Handler) CreateInvoice(w http.ResponseWriter, r *http.Request) {
	companyID, ok := middlewares.GetCompanyIDFromContext(r.Context())
	if !ok {
		http.Error(w, "No autorizado", http.StatusUnauthorized)
		return
	}

	var inv Invoice
	if err := json.NewDecoder(r.Body).Decode(&inv); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	inv.CompanyID = companyID

	if err := h.service.SaveInvoice(r.Context(), &inv); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.notify(r.Context(), notifications.CreateNotificationRequest{
		ProjectID:  &inv.ProjectID,
		EntityType: "INVOICE",
		EntityID:   &inv.ID,
		Type:       "INVOICE_CREATED",
		Priority:   notifications.PriorityMedium,
		Title:      "Nueva factura registrada",
		Message:    fmt.Sprintf("Se registró una nueva factura (%s) por $%.2f.", inv.InvoiceNumber, inv.TotalAmount),
		LinkToUI:   strPtr("/dashboard/projects/" + inv.ProjectID + "/invoices"),
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(inv)
}

func (h *Handler) PostPayment(w http.ResponseWriter, r *http.Request) {
	companyID, ok := middlewares.GetCompanyIDFromContext(r.Context())
	if !ok {
		http.Error(w, "No autorizado", http.StatusUnauthorized)
		return
	}

	var p Payment
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	p.CompanyID = companyID

	if err := h.service.ProcessPayment(r.Context(), &p); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.notify(r.Context(), notifications.CreateNotificationRequest{
		ProjectID:   &p.ProjectID,
		EntityType:  "PAYMENT",
		EntityID:    &p.ID,
		Type:        "PAYMENT_CREATED",
		Priority:    notifications.PriorityHigh,
		Title:       "Nuevo pago registrado",
		Message:     fmt.Sprintf("Se registró un pago de $%.2f.", p.Amount),
		LinkToUI:    strPtr("/dashboard/projects/" + p.ProjectID + "/invoices"),
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(p)
}

func (h *Handler) UpdateInvoice(w http.ResponseWriter, r *http.Request) {
	companyID, ok := middlewares.GetCompanyIDFromContext(r.Context())
	if !ok {
		http.Error(w, "No autorizado", http.StatusUnauthorized)
		return
	}

	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "Falta el parámetro id", http.StatusBadRequest)
		return
	}

	var req UpdateInvoiceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	if err := h.service.UpdateInvoice(r.Context(), companyID, id, req); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.notify(r.Context(), notifications.CreateNotificationRequest{
		EntityType: "INVOICE",
		EntityID:   &id,
		Type:       "INVOICE_UPDATED",
		Priority:   notifications.PriorityMedium,
		Title:      "Factura actualizada",
		Message:    "Se actualizó una factura.",
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Factura actualizada"})
}

func (h *Handler) DeleteInvoice(w http.ResponseWriter, r *http.Request) {
	companyID, ok := middlewares.GetCompanyIDFromContext(r.Context())
	if !ok {
		http.Error(w, "No autorizado", http.StatusUnauthorized)
		return
	}

	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "Falta el parámetro id", http.StatusBadRequest)
		return
	}

	if err := h.service.DeleteInvoice(r.Context(), companyID, id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.notify(r.Context(), notifications.CreateNotificationRequest{
		EntityType: "INVOICE",
		EntityID:   &id,
		Type:       "INVOICE_DELETED",
		Priority:   notifications.PriorityLow,
		Title:      "Factura eliminada",
		Message:    "Se eliminó una factura.",
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Factura eliminada"})
}

func (h *Handler) GetInvoices(w http.ResponseWriter, r *http.Request) {
	companyID, ok := middlewares.GetCompanyIDFromContext(r.Context())
	if !ok {
		http.Error(w, "No autorizado", http.StatusUnauthorized)
		return
	}

	projectID := r.PathValue("project_id")
	if projectID == "" {
		http.Error(w, "Falta project_id en la ruta", http.StatusBadRequest)
		return
	}

	invoices, err := h.service.GetProjectInvoices(r.Context(), companyID, projectID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(invoices)
}

func (h *Handler) GetInvoiceByID(w http.ResponseWriter, r *http.Request) {
	companyID, ok := middlewares.GetCompanyIDFromContext(r.Context())
	if !ok {
		http.Error(w, "No autorizado", http.StatusUnauthorized)
		return
	}

	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "Falta el id de la factura", http.StatusBadRequest)
		return
	}

	inv, err := h.service.GetInvoiceByID(r.Context(), companyID, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(inv)
}

func (h *Handler) GetPayments(w http.ResponseWriter, r *http.Request) {
	companyID, ok := middlewares.GetCompanyIDFromContext(r.Context())
	if !ok {
		http.Error(w, "No autorizado", http.StatusUnauthorized)
		return
	}

	invoiceID := r.PathValue("invoice_id")
	if invoiceID == "" {
		http.Error(w, "Falta el id de la factura", http.StatusBadRequest)
		return
	}

	payments, err := h.service.GetPayments(r.Context(), companyID, invoiceID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(payments)
}

func (h *Handler) CancelInvoice(w http.ResponseWriter, r *http.Request) {
	companyID, ok := middlewares.GetCompanyIDFromContext(r.Context())
	if !ok {
		http.Error(w, "No autorizado", http.StatusUnauthorized)
		return
	}

	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "Falta el parámetro id", http.StatusBadRequest)
		return
	}

	if err := h.service.CancelInvoice(r.Context(), companyID, id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.notify(r.Context(), notifications.CreateNotificationRequest{
		EntityType: "INVOICE",
		EntityID:   &id,
		Type:       "INVOICE_CANCELLED",
		Priority:   notifications.PriorityHigh,
		Title:      "Factura cancelada",
		Message:    "Se canceló una factura.",
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Factura cancelada"})
}
