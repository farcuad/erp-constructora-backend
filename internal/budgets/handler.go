package budgets

import (
	"encoding/json"
	"net/http"

	"erp-constructora/internal/users"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	companyId, ok := users.GetCompanyIDFromContext(r.Context())
	if !ok || companyId == "" {
		http.Error(w, "No autorizado: ID de empresa ausente", http.StatusUnauthorized)
		return
	}

	userID, ok := users.GetUserIDFromContext(r.Context())
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

func (h *Handler) GetBudgetsByProjectID(w http.ResponseWriter, r *http.Request) {
	companyId, ok := users.GetCompanyIDFromContext(r.Context())
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
