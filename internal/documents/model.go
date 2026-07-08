package documents

import (
	"time"
)

type DocumentType struct {
	ID          string    `json:"id"`
	CompanyID   string    `json:"company_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

type UpdateDocumentTypeRequest struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
}

type UpdateDocumentRequest struct {
	Title          *string `json:"title,omitempty"`
	Description    *string `json:"description,omitempty"`
	DocumentTypeID *string `json:"document_type_id,omitempty"`
}

type Document struct {
	ID             string            `json:"id"`
	CompanyID      string            `json:"company_id"`
	ProjectID      string            `json:"project_id"`
	DocumentTypeID string            `json:"document_type_id"`
	Title          string            `json:"title"`
	Description    string            `json:"description"`
	CurrentVersion int               `json:"current_version"`
	Status         string            `json:"status"`
	CreatedAt      time.Time         `json:"created_at"`
	UpdatedAt      time.Time         `json:"updated_at"`
	Versions       []DocumentVersion `json:"versions,omitempty"`
}

type DocumentVersion struct {
	ID            string    `json:"id"`
	CompanyID     string    `json:"company_id"`
	DocumentID    string    `json:"document_id"`
	VersionNumber int       `json:"version_number"`
	FileURL       string    `json:"file_url"`
	FileSize      int64     `json:"file_size"`
	FileExtension string    `json:"file_extension"`
	ChangeLog     string    `json:"change_log"`
	UserID        string    `json:"user_id"`
	CreatedAt     time.Time `json:"created_at"`
}
