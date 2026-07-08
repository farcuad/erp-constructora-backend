package expense

import (
	"encoding/json"
	"net/http"

	"erp-constructora/internal/middlewares"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	companyID, ok := middlewares.GetCompanyIDFromContext(r.Context())
	if !ok || companyID == "" {
		http.Error(w, "No autorizado", http.StatusUnauthorized)
		return
	}
	userID, ok := middlewares.GetUserIDFromContext(r.Context())
	if !ok || userID == "" {
		http.Error(w, "No autorizado", http.StatusUnauthorized)
		return
	}

	var req CreateExpenseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Payload inválido", http.StatusBadRequest)
		return
	}

	expense, err := h.service.RegisterExpense(r.Context(), companyID, userID, &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(expense)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	companyID, ok := middlewares.GetCompanyIDFromContext(r.Context())
	if !ok || companyID == "" {
		http.Error(w, "No autorizado", http.StatusUnauthorized)
		return
	}

	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "ID del gasto es requerido", http.StatusBadRequest)
		return
	}

	var req UpdateExpenseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Payload inválido", http.StatusBadRequest)
		return
	}

	expense, err := h.service.UpdateExpense(r.Context(), companyID, id, &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(expense)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	companyID, ok := middlewares.GetCompanyIDFromContext(r.Context())
	if !ok || companyID == "" {
		http.Error(w, "No autorizado", http.StatusUnauthorized)
		return
	}

	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "ID del gasto es requerido", http.StatusBadRequest)
		return
	}

	err := h.service.DeleteExpense(r.Context(), companyID, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "recurso eliminado"})
}

func (h *Handler) GetByProject(w http.ResponseWriter, r *http.Request) {
	companyID, ok := middlewares.GetCompanyIDFromContext(r.Context())
	if !ok || companyID == "" {
		http.Error(w, "No autorizado", http.StatusUnauthorized)
		return
	}

	projectID := r.PathValue("project_id")

	if projectID == "" {
		http.Error(w, "Falta el ID del proyecto", http.StatusBadRequest)
		return
	}

	expenses, err := h.service.GetProjectExpenses(r.Context(), companyID, projectID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(expenses)
}
