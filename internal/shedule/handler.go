package schedule

import (
	"encoding/json"
	"net/http"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
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

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(t)
}

func (h *Handler) AddDependency(w http.ResponseWriter, r *http.Request) {
	var d TaskDependency
	if err := json.NewDecoder(r.Body).Decode(&d); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	if err := h.service.LinkTasks(r.Context(), &d); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(d)
}

func (h *Handler) CreateMilestone(w http.ResponseWriter, r *http.Request) {
	var m Milestone
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	if err := h.service.CreateMilestone(r.Context(), &m); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(m)
}

func (h *Handler) GetSchedule(w http.ResponseWriter, r *http.Request) {
	// Utilizando tu estilo limpio /{project_id}
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
