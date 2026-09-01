package attendance

import (
	"context"
	"encoding/json"
	"erp-constructora/internal/middlewares"
	"erp-constructora/internal/notifications"
	"erp-constructora/internal/utils"
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
		log.Printf("[NOTIFY ERROR] attendance: %v", err)
	}
}

func (h *Handler) SaveAttendance(w http.ResponseWriter, r *http.Request) {
	companyID, ok := middlewares.GetCompanyIDFromContext(r.Context())
	if !ok {
		utils.WriteUnauthorized(w)
		return
	}

	var att Attendance
	if err := json.NewDecoder(r.Body).Decode(&att); err != nil {
		utils.WriteBadRequest(w, "JSON inválido: "+err.Error())
		return
	}
	att.CompanyID = companyID

	if err := h.service.SubmitAttendance(r.Context(), &att); err != nil {
		utils.WriteBadRequest(w, utils.GetPGErrorMessage(err))
		return
	}

	h.notify(r.Context(), notifications.CreateNotificationRequest{
		ProjectID:  &att.ProjectID,
		EntityType: "ATTENDANCE",
		Type:       "ATTENDANCE_SAVED",
		Priority:   notifications.PriorityLow,
		Title:      "Asistencia registrada",
		Message:    "Se registró la asistencia del personal de la obra.",
		LinkToUI:   strPtr("/dashboard/projects/" + att.ProjectID + "/attendance"),
	})

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
		utils.WriteBadRequest(w, "El parámetro project_id en la ruta y el query 'date' son obligatorios")
		return
	}

	report, err := h.service.GetDailyReport(r.Context(), projectID, date)
	if err != nil {
		utils.WriteInternalError(w, utils.GetPGErrorMessage(err))
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
		utils.WriteBadRequest(w, "Falta el id del registro de asistencia")
		return
	}

	var req UpdateAttendanceLogRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteBadRequest(w, "JSON inválido: "+err.Error())
		return
	}

	log := AttendanceLog{
		ID:          id,
		Status:      req.Status,
		HoursWorked: req.HoursWorked,
		Notes:       req.Notes,
	}

	if err := h.service.UpdateAttendanceLog(r.Context(), &log); err != nil {
		utils.WriteBadRequest(w, utils.GetPGErrorMessage(err))
		return
	}

	h.notify(r.Context(), notifications.CreateNotificationRequest{
		EntityType: "ATTENDANCE_LOG",
		EntityID:   &id,
		Type:       "ATTENDANCE_UPDATED",
		Priority:   notifications.PriorityLow,
		Title:      "Registro de asistencia actualizado",
		Message:    "Se actualizó un registro de asistencia.",
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(log)
}

func (h *Handler) DeleteAttendance(w http.ResponseWriter, r *http.Request) {
	companyID, ok := middlewares.GetCompanyIDFromContext(r.Context())
	if !ok {
		utils.WriteUnauthorized(w)
		return
	}

	id := r.PathValue("id")
	if id == "" {
		utils.WriteBadRequest(w, "Falta el id de la asistencia")
		return
	}

	if err := h.service.DeleteAttendance(r.Context(), companyID, id); err != nil {
		utils.WriteInternalError(w, utils.GetPGErrorMessage(err))
		return
	}

	h.notify(r.Context(), notifications.CreateNotificationRequest{
		EntityType: "ATTENDANCE",
		EntityID:   &id,
		Type:       "ATTENDANCE_DELETED",
		Priority:   notifications.PriorityLow,
		Title:      "Asistencia eliminada",
		Message:    "Se eliminó un registro de asistencia.",
	})

	w.WriteHeader(http.StatusNoContent)
}
