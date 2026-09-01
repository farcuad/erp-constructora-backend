package financialdashboard

import (
	"encoding/json"
	"erp-constructora/internal/users"
	"erp-constructora/internal/utils"
	"net/http"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) GetSummary(w http.ResponseWriter, r *http.Request) {
	companyID, ok := users.GetCompanyIDFromContext(r.Context())
	if !ok {
		utils.WriteUnauthorized(w)
		return
	}

	projectID := r.PathValue("project_id")
	if projectID == "" {
		utils.WriteBadRequest(w, "El parámetro project_id es requerido")
		return
	}

	kpis, err := h.service.GetDashboardKPIs(r.Context(), companyID, projectID)
	if err != nil {
		utils.WriteInternalError(w, utils.GetPGErrorMessage(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(kpis)
}
