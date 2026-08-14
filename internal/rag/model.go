package rag

import (
	"time"
)

type ChunkResult struct {
	DocumentID string
	Content    string
	Score      float64
}

// ChatRequest es la estructura que recibe Go desde la burbuja de React
type ChatRequest struct {
	Question  string  `json:"question"`
	ProjectID *string `json:"project_id,omitempty"` // Opcional: Para filtrar la búsqueda a un proyecto específico
}

// DocumentSource representa los metadatos del documento usado como evidencia
type DocumentSource struct {
	DocumentID string `json:"document_id"`
	Title      string `json:"title"`
}

// ChatResponse es lo que le devolvemos a la app en React
type ChatResponse struct {
	Answer    string           `json:"answer"`
	Sources   []DocumentSource `json:"sources"`
	CreatedAt time.Time        `json:"created_at"`
}
