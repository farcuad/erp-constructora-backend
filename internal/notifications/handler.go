package notifications

import (
	"encoding/json"
	"net/http"

	"erp-constructora/internal/middlewares"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Configurar orígenes según tu dominio en producción
	},
}

type Handler struct {
	service *Service
	hub     *WSHub
}

func NewHandler(service *Service, hub *WSHub) *Handler {
	return &Handler{service: service, hub: hub}
}

// HandleWS gestiona el apretón de manos para conectar React con el backend mediante WebSockets
func (h *Handler) HandleWS(w http.ResponseWriter, r *http.Request) {
	companyID, okCompany := middlewares.GetCompanyIDFromContext(r.Context())
	userID, okUser := middlewares.GetUserIDFromContext(r.Context())

	if !okCompany || !okUser {
		http.Error(w, "No autorizado", http.StatusUnauthorized)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	h.hub.Register(companyID, userID, conn)

	// Mantener la conexión abierta escuchando desconexiones
	go func() {
		defer h.hub.Unregister(companyID, userID)
		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				break
			}
		}
	}()
}

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

// RegisterPushToken registra el token FCM del dispositivo móvil (Flutter lo llama tras el login)
func (h *Handler) RegisterPushToken(w http.ResponseWriter, r *http.Request) {
	userID, ok := middlewares.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "No autorizado", http.StatusUnauthorized)
		return
	}

	var req struct {
		Token    string `json:"token"`
		Platform string `json:"platform"` // 'android' / 'ios'
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Token == "" {
		http.Error(w, "Se requiere el campo 'token'", http.StatusBadRequest)
		return
	}

	if err := h.service.RegisterPushToken(r.Context(), userID, req.Token, req.Platform); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Token de push registrado"})
}

// UnregisterPushToken elimina el token FCM al cerrar sesión en la app móvil
func (h *Handler) UnregisterPushToken(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Token == "" {
		http.Error(w, "Se requiere el campo 'token'", http.StatusBadRequest)
		return
	}

	if err := h.service.UnregisterPushToken(r.Context(), req.Token); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Token de push eliminado"})
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
