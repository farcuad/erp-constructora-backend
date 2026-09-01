package schedule

import (
	"context"
	"encoding/json"
	"erp-constructora/internal/notifications"
	"erp-constructora/internal/utils"
	"fmt"
	"log"
	"net/http"
)

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
		log.Printf("[NOTIFY ERROR] schedule: %v", err)
	}
}

func (h *Handler) CreateTask(w http.ResponseWriter, r *http.Request) {
	var t Task
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		utils.WriteBadRequest(w, "JSON inválido")
		return
	}

	if err := h.service.CreateTask(r.Context(), &t); err != nil {
		utils.WriteBadRequest(w, utils.GetPGErrorMessage(err))
		return
	}

	h.notify(r.Context(), notifications.CreateNotificationRequest{
		ProjectID:  &t.ProjectID,
		EntityType: "SCHEDULE_TASK",
		EntityID:   &t.ID,
		Type:       "TASK_CREATED",
		Priority:   notifications.PriorityMedium,
		Title:      "Nueva tarea añadida",
		Message:    fmt.Sprintf("Se añadió una nueva tarea: %s, revisa las nuevas tareas.", t.Name),
		LinkToUI:   strPtr("/dashboard/projects/" + t.ProjectID + "/schedule"),
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(t)
}

func (h *Handler) GetSchedule(w http.ResponseWriter, r *http.Request) {
	projectID := r.PathValue("project_id")
	if projectID == "" {
		utils.WriteBadRequest(w, "Falta project_id en la ruta")
		return
	}

	tasks, err := h.service.GetProjectTasks(r.Context(), projectID)
	if err != nil {
		utils.WriteInternalError(w, utils.GetPGErrorMessage(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tasks)
}

func (h *Handler) UpdateTask(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		utils.WriteBadRequest(w, "Falta el id de la tarea")
		return
	}

	var t Task
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		utils.WriteBadRequest(w, "JSON inválido")
		return
	}

	t.ID = id

	if err := h.service.UpdateTask(r.Context(), &t); err != nil {
		utils.WriteBadRequest(w, utils.GetPGErrorMessage(err))
		return
	}

	h.notify(r.Context(), notifications.CreateNotificationRequest{
		ProjectID:  &t.ProjectID,
		EntityType: "SCHEDULE_TASK",
		EntityID:   &t.ID,
		Type:       "TASK_UPDATED",
		Priority:   notifications.PriorityMedium,
		Title:      "Tarea actualizada",
		Message:    fmt.Sprintf("Se actualizó la tarea: %s.", t.Name),
		LinkToUI:   strPtr("/dashboard/projects/" + t.ProjectID + "/schedule"),
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(t)
}

func (h *Handler) DeleteTask(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		utils.WriteBadRequest(w, "Falta el id de la tarea")
		return
	}

	if err := h.service.DeleteTask(r.Context(), id); err != nil {
		utils.WriteInternalError(w, utils.GetPGErrorMessage(err))
		return
	}

	h.notify(r.Context(), notifications.CreateNotificationRequest{
		EntityType: "SCHEDULE_TASK",
		EntityID:   &id,
		Type:       "TASK_DELETED",
		Priority:   notifications.PriorityMedium,
		Title:      "Tarea eliminada",
		Message:    "Se eliminó una tarea del cronograma.",
	})

	w.WriteHeader(http.StatusNoContent)
}

// strPtr devuelve un puntero al string, útil para los campos *string del payload de notificación
func strPtr(s string) *string {
	return &s
}
