package audit

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"erp-constructora/internal/notifications"
	"erp-constructora/internal/users"
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
		log.Printf("[NOTIFY ERROR] audit_logs: %v", err)
	}
}

// helper para extraer la IP real del cliente detrás de proxies comunes (como Supabase/Cloudflare)
func getClientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		return strings.TrimSpace(parts[0])
	}
	return strings.Split(r.RemoteAddr, ":")[0]
}

func (h *Handler) GetCompanyLogs(w http.ResponseWriter, r *http.Request) {
	companyID, ok := users.GetCompanyIDFromContext(r.Context())
	if !ok {
		http.Error(w, "No autorizado", http.StatusUnauthorized)
		return
	}

	list, err := h.service.FetchLogs(r.Context(), companyID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(list)
}

// Nota: Normalmente no expones un "POST" público para auditorías, sino que otros módulos
// llaman internamente al Service. Pero si tu Frontend (React/Flutter) necesita auditar eventos puros
// del cliente (ej. "EXPORT_FINANCIAL_EXCEL"), puedes usar este handler:
func (h *Handler) CreateLog(w http.ResponseWriter, r *http.Request) {
	companyID, ok := users.GetCompanyIDFromContext(r.Context())
	userID, okUser := users.GetUserIDFromContext(r.Context())
	if !ok || !okUser {
		http.Error(w, "No autorizado", http.StatusUnauthorized)
		return
	}

	var req CreateAuditRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	logEntry := AuditLog{
		CompanyID: companyID,
		UserID:    userID,
		Action:    req.Action,
		TableName: req.TableName,
		RowID:     req.RowID,
		IPAddress: getClientIP(r),
		OldValues: req.OldValues,
		NewValues: req.NewValues,
	}

	if err := h.service.RegisterLog(r.Context(), &logEntry); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.notify(r.Context(), notifications.CreateNotificationRequest{
		EntityType: "AUDIT_LOG",
		Type:       req.Action,
		Priority:   notifications.PriorityLow,
		Title:      "Nueva actividad registrada",
		Message:    fmt.Sprintf("Se registró una actividad: %s en %s.", req.Action, req.TableName),
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(logEntry)
}
