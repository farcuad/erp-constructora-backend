package notifications

import (
	"encoding/json"
	"time"
)

type NotificationPriority string

const (
	PriorityLow      NotificationPriority = "low"
	PriorityMedium   NotificationPriority = "medium"
	PriorityHigh     NotificationPriority = "high"
	PriorityCritical NotificationPriority = "critical"
)

type Notification struct {
	ID         string               `json:"id"`
	CompanyID  string               `json:"company_id"`
	ProjectID  *string              `json:"project_id,omitempty"` // Puntero para soportar NULL
	EntityType string               `json:"entity_type"`          // Ej: 'PURCHASE_ORDER', 'BUDGET_ITEM', 'DAILY_REPORT'
	EntityID   *string              `json:"entity_id,omitempty"`
	Type       string               `json:"type"`     // Ej: 'APPROVAL_REQUIRED', 'BUDGET_OVERRUN'
	Priority   NotificationPriority `json:"priority"` // 'low', 'medium', 'high', 'critical'
	Title      string               `json:"title"`
	Message    string               `json:"message"`
	LinkToUI   *string              `json:"link_to_ui,omitempty"` // Cambio a puntero para soportar NULL nativo
	Metadata   json.RawMessage      `json:"metadata,omitempty"`   // Flexibilidad JSONB
	CreatedAt  time.Time            `json:"created_at"`
	IsRead     bool                 `json:"is_read"` // Campo calculado
}

type CreateNotificationRequest struct {
	ProjectID   *string              `json:"project_id,omitempty"`
	EntityType  string               `json:"entity_type"`
	EntityID    *string              `json:"entity_id,omitempty"`
	Type        string               `json:"type"`
	Priority    NotificationPriority `json:"priority"`
	Title       string               `json:"title"`
	Message     string               `json:"message"`
	LinkToUI    *string              `json:"link_to_ui,omitempty"`
	Metadata    map[string]any       `json:"metadata,omitempty"`
	TargetUsers []string             `json:"target_users"`
}
