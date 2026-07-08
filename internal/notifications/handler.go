package notifications

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

func (h *Handler) CreateNotifications(w http.ResponseWriter, r *http.Request) {
	companyID, ok := middlewares.GetCompanyIDFromContext(r.Context())
	if !ok {
		http.Error(w, "No autorizado", http.StatusUnauthorized)
		return
	}

	var req CreateNotificationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	if req.Title == "" || req.Message == "" {
		http.Error(w, "El título y el mensaje son requeridos", http.StatusBadRequest)
		return
	}
	if len(req.TargetUsers) == 0 {
		http.Error(w, "Debe especificar al menos un usuario en target_users", http.StatusBadRequest)
		return
	}

	notification := Notification{
		CompanyID: companyID,
		ProjectID: req.ProjectID,
		Title:     req.Title,
		Message:   req.Message,
		LinkToUI:  req.LinkToUI,
	}

	err := h.service.DispatchNotification(r.Context(), &notification, req.TargetUsers)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(notification)
}

// GetMyNotifications responde con la bandeja de entrada del usuario autenticado
func (h *Handler) GetMyNotifications(w http.ResponseWriter, r *http.Request) {
	companyID, ok := middlewares.GetCompanyIDFromContext(r.Context())
	userID, okUser := middlewares.GetUserIDFromContext(r.Context())
	if !ok || !okUser {
		http.Error(w, "No autorizado", http.StatusUnauthorized)
		return
	}

	list, err := h.service.FetchMyNotifications(r.Context(), companyID, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(list)
}

// MarkRead Endpoint para cuando el usuario hace clic sobre la notificación en React/Flutter
func (h *Handler) MarkRead(w http.ResponseWriter, r *http.Request) {
	companyID, ok := middlewares.GetCompanyIDFromContext(r.Context())
	userID, okUser := middlewares.GetUserIDFromContext(r.Context())
	if !ok || !okUser {
		http.Error(w, "No autorizado", http.StatusUnauthorized)
		return
	}

	notificationID := r.PathValue("notification_id")
	if notificationID == "" {
		http.Error(w, "Falta el parámetro notification_id", http.StatusBadRequest)
		return
	}

	err := h.service.ReadNotification(r.Context(), companyID, notificationID, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Notificación marcada como leída"})
}

func (h *Handler) DeleteNotification(w http.ResponseWriter, r *http.Request) {
	companyID, ok := middlewares.GetCompanyIDFromContext(r.Context())
	if !ok {
		http.Error(w, "No autorizado", http.StatusUnauthorized)
		return
	}

	id := r.PathValue("notification_id")
	if id == "" {
		http.Error(w, "Falta el parámetro notification_id", http.StatusBadRequest)
		return
	}

	if err := h.service.DeleteNotification(r.Context(), companyID, id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Notificación eliminada"})
}
