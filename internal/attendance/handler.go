package attendance

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

func (h *Handler) SaveAttendance(w http.ResponseWriter, r *http.Request) {
	companyID, ok := middlewares.GetCompanyIDFromContext(r.Context())
	if !ok {
		http.Error(w, "No autorizado", http.StatusUnauthorized)
		return
	}

	var att Attendance
	if err := json.NewDecoder(r.Body).Decode(&att); err != nil {
		http.Error(w, "JSON inválido: "+err.Error(), http.StatusBadRequest)
		return
	}
	att.CompanyID = companyID

	if err := h.service.SubmitAttendance(r.Context(), &att); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(att)
}

func (h *Handler) GetAttendance(w http.ResponseWriter, r *http.Request) {
	// 1. Extraer el project_id directamente de la ruta limpia
	projectID := r.PathValue("project_id")

	// 2. Extraer la fecha desde el query string (Ej: /attendance/uuid-proyecto?date=2026-07-02)
	date := r.URL.Query().Get("date")

	if projectID == "" || date == "" {
		http.Error(w, "El parámetro project_id en la ruta y el query 'date' son obligatorios", http.StatusBadRequest)
		return
	}

	report, err := h.service.GetDailyReport(r.Context(), projectID, date)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if report == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"message": "No se encontró registro de asistencia para este proyecto en la fecha especificada"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(report)
}

func (h *Handler) UpdateAttendanceLog(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "Falta el id del registro de asistencia", http.StatusBadRequest)
		return
	}

	var req UpdateAttendanceLogRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "JSON inválido: "+err.Error(), http.StatusBadRequest)
		return
	}

	log := AttendanceLog{
		ID:          id,
		Status:      req.Status,
		HoursWorked: req.HoursWorked,
		Notes:       req.Notes,
	}

	if err := h.service.UpdateAttendanceLog(r.Context(), &log); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(log)
}

func (h *Handler) DeleteAttendance(w http.ResponseWriter, r *http.Request) {
	companyID, ok := middlewares.GetCompanyIDFromContext(r.Context())
	if !ok {
		http.Error(w, "No autorizado", http.StatusUnauthorized)
		return
	}

	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "Falta el id de la asistencia", http.StatusBadRequest)
		return
	}

	if err := h.service.DeleteAttendance(r.Context(), companyID, id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
