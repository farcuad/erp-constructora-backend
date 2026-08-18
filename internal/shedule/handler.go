package schedule

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"erp-constructora/internal/notifications"
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
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	if err := h.service.CreateTask(r.Context(), &t); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
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
		http.Error(w, "Falta project_id en la ruta", http.StatusBadRequest)
		return
	}

	tasks, err := h.service.GetProjectTasks(r.Context(), projectID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tasks)
}

func (h *Handler) UpdateTask(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "Falta el id de la tarea", http.StatusBadRequest)
		return
	}

	var t Task
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	t.ID = id

	if err := h.service.UpdateTask(r.Context(), &t); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
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
		http.Error(w, "Falta el id de la tarea", http.StatusBadRequest)
		return
	}

	if err := h.service.DeleteTask(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
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
