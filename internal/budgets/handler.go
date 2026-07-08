package budgets

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
	companyId, ok := middlewares.GetCompanyIDFromContext(r.Context())
	if !ok || companyId == "" {
		http.Error(w, "No autorizado: ID de empresa ausente", http.StatusUnauthorized)
		return
	}

	userID, ok := middlewares.GetUserIDFromContext(r.Context())
	if !ok || userID == "" {
		http.Error(w, "No autorizado: ID de usuario ausente", http.StatusUnauthorized)
		return
	}

	var req CreateBudgetWithItemsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Payload inválido", http.StatusBadRequest)
		return
	}

	budget, err := h.service.CreateInitialBudget(r.Context(), companyId, userID, &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(budget)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	companyId, ok := middlewares.GetCompanyIDFromContext(r.Context())
	if !ok || companyId == "" {
		http.Error(w, "No autorizado: ID de empresa ausente", http.StatusUnauthorized)
		return
	}

	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "ID del presupuesto es requerido", http.StatusBadRequest)
		return
	}

	var req UpdateBudgetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Payload inválido", http.StatusBadRequest)
		return
	}

	budget, err := h.service.UpdateBudget(r.Context(), companyId, id, &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(budget)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	companyId, ok := middlewares.GetCompanyIDFromContext(r.Context())
	if !ok || companyId == "" {
		http.Error(w, "No autorizado: ID de empresa ausente", http.StatusUnauthorized)
		return
	}

	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "ID del presupuesto es requerido", http.StatusBadRequest)
		return
	}

	err := h.service.DeleteBudget(r.Context(), companyId, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "recurso eliminado"})
}

func (h *Handler) GetBudgetsByProjectID(w http.ResponseWriter, r *http.Request) {
	companyId, ok := middlewares.GetCompanyIDFromContext(r.Context())
	if !ok || companyId == "" {
		http.Error(w, "No autorizado: ID de empresa ausente", http.StatusUnauthorized)
		return
	}
	projectId := r.PathValue("project_id")
	if projectId == "" {
		http.Error(w, "ID de proyecto es requerido", http.StatusBadRequest)
		return
	}
	budgets, err := h.service.GetBudgetsProjectID(r.Context(), companyId, projectId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(budgets)
}
