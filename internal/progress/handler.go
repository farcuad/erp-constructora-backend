package progress

import (
	"context"
	"database/sql"
	"encoding/json"
	"erp-constructora/internal/middlewares"
	"erp-constructora/internal/notifications"
	"erp-constructora/internal/utils"
	"errors"
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
		log.Printf("[NOTIFY ERROR] progress: %v", err)
	}
}

func (h *Handler) CreateDailyReport(w http.ResponseWriter, r *http.Request) {
	var report DailyReport
	if err := json.NewDecoder(r.Body).Decode(&report); err != nil {
		utils.WriteBadRequest(w, "JSON inválido")
		return
	}

	companyID, ok := middlewares.GetCompanyIDFromContext(r.Context())
	if !ok {
		utils.WriteUnauthorized(w)
		return
	}
	report.CompanyID = companyID

	userID, ok := middlewares.GetUserIDFromContext(r.Context())
	if !ok {
		utils.WriteUnauthorized(w)
		return
	}
	report.UserID = userID

	if report.ProjectID == "" {
		utils.WriteBadRequest(w, "project_id es requerido")
		return
	}

	err := h.service.SaveDailyProgress(r.Context(), &report)
	if err != nil {
		utils.WriteInternalError(w, utils.GetPGErrorMessage(err))
		return
	}

	h.notify(r.Context(), notifications.CreateNotificationRequest{
		ProjectID:  &report.ProjectID,
		EntityType: "DAILY_REPORT",
		Type:       "REPORT_CREATED",
		Priority:   notifications.PriorityMedium,
		Title:      "Nuevo reporte diario de obra",
		Message:    "Se registró un nuevo reporte diario de obra.",
		LinkToUI:   strPtr("/dashboard/projects/" + report.ProjectID + "/progress"),
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(report)
}

func (h *Handler) GetDailyReport(w http.ResponseWriter, r *http.Request) {
	companyID, ok := middlewares.GetCompanyIDFromContext(r.Context())
	if !ok {
		utils.WriteUnauthorized(w)
		return
	}

	projectID := r.PathValue("project_id")
	date := r.URL.Query().Get("date")

	if projectID == "" || date == "" {
		utils.WriteBadRequest(w, "El parámetro project_id en la ruta y el query 'date' son obligatorios")
		return
	}

	report, err := h.service.GetDailyReport(r.Context(), companyID, projectID, date)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"message": "No se encontró reporte para este proyecto en la fecha especificada"})
			return
		}
		utils.WriteInternalError(w, utils.GetPGErrorMessage(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(report)
}

func (h *Handler) UpdateDailyReport(w http.ResponseWriter, r *http.Request) {
	companyID, ok := middlewares.GetCompanyIDFromContext(r.Context())
	if !ok {
		utils.WriteUnauthorized(w)
		return
	}

	id := r.PathValue("id")
	if id == "" {
		utils.WriteBadRequest(w, "Falta el parámetro id")
		return
	}

	var req UpdateDailyReportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteBadRequest(w, "JSON inválido")
		return
	}

	if err := h.service.UpdateDailyReport(r.Context(), companyID, id, req); err != nil {
		utils.WriteInternalError(w, utils.GetPGErrorMessage(err))
		return
	}

	h.notify(r.Context(), notifications.CreateNotificationRequest{
		EntityType: "DAILY_REPORT",
		EntityID:   &id,
		Type:       "REPORT_UPDATED",
		Priority:   notifications.PriorityMedium,
		Title:      "Reporte diario actualizado",
		Message:    "Se actualizó un reporte diario de obra.",
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Reporte diario actualizado"})
}

func (h *Handler) DeleteDailyReport(w http.ResponseWriter, r *http.Request) {
	companyID, ok := middlewares.GetCompanyIDFromContext(r.Context())
	if !ok {
		utils.WriteUnauthorized(w)
		return
	}

	id := r.PathValue("id")
	if id == "" {
		utils.WriteBadRequest(w, "Falta el parámetro id")
		return
	}

	if err := h.service.DeleteDailyReport(r.Context(), companyID, id); err != nil {
		utils.WriteInternalError(w, utils.GetPGErrorMessage(err))
		return
	}

	h.notify(r.Context(), notifications.CreateNotificationRequest{
		EntityType: "DAILY_REPORT",
		EntityID:   &id,
		Type:       "REPORT_DELETED",
		Priority:   notifications.PriorityMedium,
		Title:      "Reporte diario eliminado",
		Message:    "Se eliminó un reporte diario de obra.",
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Reporte diario eliminado"})
}
