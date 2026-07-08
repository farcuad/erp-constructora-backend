package progress

import (
	"encoding/json"
	"net/http"

	"erp-constructora/internal/middlewares"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) CreateDailyReport(w http.ResponseWriter, r *http.Request) {
	var report DailyReport
	if err := json.NewDecoder(r.Body).Decode(&report); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	companyID, ok := middlewares.GetCompanyIDFromContext(r.Context())
	if !ok {
		http.Error(w, "No autorizado", http.StatusUnauthorized)
		return
	}
	report.CompanyID = companyID

	userID, ok := middlewares.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "No autorizado", http.StatusUnauthorized)
		return
	}
	report.UserID = userID

	if report.ProjectID == "" {
		http.Error(w, "project_id es requerido", http.StatusBadRequest)
		return
	}

	err := h.service.SaveDailyProgress(r.Context(), &report)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(report)
}

func (h *Handler) GetDailyReport(w http.ResponseWriter, r *http.Request) {
	companyID, ok := middlewares.GetCompanyIDFromContext(r.Context())
	if !ok {
		http.Error(w, "No autorizado", http.StatusUnauthorized)
		return
	}

	projectID := r.PathValue("project_id")
	date := r.URL.Query().Get("date")

	if projectID == "" || date == "" {
		http.Error(w, "El parámetro project_id en la ruta y el query 'date' son obligatorios", http.StatusBadRequest)
		return
	}

	report, err := h.service.GetDailyReport(r.Context(), companyID, projectID, date)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if report == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"message": "No se encontró reporte para este proyecto en la fecha especificada"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(report)
}

func (h *Handler) UpdateDailyReport(w http.ResponseWriter, r *http.Request) {
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

	var req UpdateDailyReportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	if err := h.service.UpdateDailyReport(r.Context(), companyID, id, req); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Reporte diario actualizado"})
}

func (h *Handler) DeleteDailyReport(w http.ResponseWriter, r *http.Request) {
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

	if err := h.service.DeleteDailyReport(r.Context(), companyID, id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Reporte diario eliminado"})
}
