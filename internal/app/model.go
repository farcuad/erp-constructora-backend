package app

import "time"

// Release representa una versión publicada de la app (el .apk compilado).
type Release struct {
	ID          string    `json:"id"`
	Version     string    `json:"version"`
	AppURL      string    `json:"app_url"`
	Description string    `json:"description,omitempty"`
	FileSize    int64     `json:"file_size,omitempty"`
	Checksum    string    `json:"checksum,omitempty"`
	IsMandatory bool      `json:"is_mandatory"`
	CreatedAt   time.Time `json:"created_at"`
}

// CreateReleaseRequest es el payload del POST /app/releases (lo dispara el proceso de compilación).
type CreateReleaseRequest struct {
	Version     string `json:"version"`
	AppURL      string `json:"app_url"`
	Description string `json:"description,omitempty"`
	FileSize    int64  `json:"file_size,omitempty"`
	Checksum    string `json:"checksum,omitempty"`
	IsMandatory bool   `json:"is_mandatory"`
}
