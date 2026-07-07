package notifications

import (
	"time"
)

type Notification struct {
	ID        string    `json:"id"`
	CompanyID string    `json:"company_id"`
	ProjectID *string   `json:"project_id,omitempty"` // Puntero para soportar NULL
	Title     string    `json:"title"`
	Message   string    `json:"message"`
	LinkToUI  *string   `json:"link_to_ui,omitempty"` // Cambio a puntero para soportar NULL nativo
	CreatedAt time.Time `json:"created_at"`
	IsRead    bool      `json:"is_read"` // Campo calculado
}

type CreateNotificationRequest struct {
	ProjectID   *string  `json:"project_id"`
	Title       string   `json:"title"`
	Message     string   `json:"message"`
	LinkToUI    *string  `json:"link_to_ui"`
	TargetUsers []string `json:"target_users"`
}
