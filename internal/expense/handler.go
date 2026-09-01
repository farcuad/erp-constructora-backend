package expense

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
	service  Service
	notifier notifications.Notifier
}

func NewHandler(service Service, notifier notifications.Notifier) *Handler {
	return &Handler{service: service, notifier: notifier}
}

// notify emite una notificación a toda la empresa (el actor queda excluido por NotifyFromContext)
func (h *Handler) notify(ctx context.Context, req notifications.CreateNotificationRequest) {
	if h.notifier == nil {
		return
	}
	if err := h.notifier.NotifyFromContext(ctx, req); err != nil {
		log.Printf("[NOTIFY ERROR] expense: %v", err)
	}
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	companyID, ok := middlewares.GetCompanyIDFromContext(r.Context())
	if !ok || companyID == "" {
		utils.WriteUnauthorized(w)
		return
	}
	userID, ok := middlewares.GetUserIDFromContext(r.Context())
	if !ok || userID == "" {
		utils.WriteUnauthorized(w)
		return
	}

	var req CreateExpenseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteBadRequest(w, "Payload inválido")
		return
	}

	expense, err := h.service.RegisterExpense(r.Context(), companyID, userID, &req)
	if err != nil {
		utils.WriteInternalError(w, utils.GetPGErrorMessage(err))
		return
	}

	h.notify(r.Context(), notifications.CreateNotificationRequest{
		ProjectID:  &req.ProjectID,
		EntityType: "EXPENSE",
		EntityID:   &expense.ID,
		Type:       "EXPENSE_CREATED",
		Priority:   notifications.PriorityMedium,
		Title:      "Nuevo gasto registrado",
		Message:    fmt.Sprintf("Se registró un gasto de $%.2f: %s.", expense.Amount, expense.Title),
		LinkToUI:   strPtr("/dashboard/projects/" + req.ProjectID + "/expenses"),
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(expense)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	companyID, ok := middlewares.GetCompanyIDFromContext(r.Context())
	if !ok || companyID == "" {
		utils.WriteUnauthorized(w)
		return
	}

	id := r.PathValue("id")
	if id == "" {
		utils.WriteBadRequest(w, "ID del gasto es requerido")
		return
	}

	var req UpdateExpenseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteBadRequest(w, "Payload inválido")
		return
	}

	expense, err := h.service.UpdateExpense(r.Context(), companyID, id, &req)
	if err != nil {
		utils.WriteInternalError(w, utils.GetPGErrorMessage(err))
		return
	}

	h.notify(r.Context(), notifications.CreateNotificationRequest{
		EntityType: "EXPENSE",
		EntityID:   &id,
		Type:       "EXPENSE_UPDATED",
		Priority:   notifications.PriorityMedium,
		Title:      "Gasto actualizado",
		Message:    "Se actualizó un gasto del proyecto.",
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(expense)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	companyID, ok := middlewares.GetCompanyIDFromContext(r.Context())
	if !ok || companyID == "" {
		utils.WriteUnauthorized(w)
		return
	}

	id := r.PathValue("id")
	if id == "" {
		utils.WriteBadRequest(w, "ID del gasto es requerido")
		return
	}

	err := h.service.DeleteExpense(r.Context(), companyID, id)
	if err != nil {
		utils.WriteInternalError(w, utils.GetPGErrorMessage(err))
		return
	}

	h.notify(r.Context(), notifications.CreateNotificationRequest{
		EntityType: "EXPENSE",
		EntityID:   &id,
		Type:       "EXPENSE_DELETED",
		Priority:   notifications.PriorityMedium,
		Title:      "Gasto eliminado",
		Message:    "Se eliminó un gasto del proyecto.",
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "recurso eliminado"})
}

func (h *Handler) GetByProject(w http.ResponseWriter, r *http.Request) {
	companyID, ok := middlewares.GetCompanyIDFromContext(r.Context())
	if !ok || companyID == "" {
		utils.WriteUnauthorized(w)
		return
	}

	projectID := r.PathValue("project_id")

	if projectID == "" {
		utils.WriteBadRequest(w, "Falta el ID del proyecto")
		return
	}

	expenses, err := h.service.GetProjectExpenses(r.Context(), companyID, projectID)
	if err != nil {
		utils.WriteInternalError(w, utils.GetPGErrorMessage(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(expenses)
}
