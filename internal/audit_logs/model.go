package audit

import (
	"encoding/json"
	"time"
)

type AuditLog struct {
	ID        string          `json:"id"`
	CompanyID string          `json:"company_id"`
	UserID    string          `json:"user_id"`
	Action    string          `json:"action"`
	TableName string          `json:"table_name"`
	RowID     *string         `json:"row_id,omitempty"`
	IPAddress string          `json:"ip_address"`
	OldValues json.RawMessage `json:"old_values,omitempty"` // json.RawMessage mapea directo a JSONB de Postgres
	NewValues json.RawMessage `json:"new_values,omitempty"`
	CreatedAt time.Time       `json:"created_at"`
}

type CreateAuditRequest struct {
	Action    string          `json:"action"`
	TableName string          `json:"table_name"`
	RowID     *string         `json:"row_id"`
	OldValues json.RawMessage `json:"old_values"`
	NewValues json.RawMessage `json:"new_values"`
}
