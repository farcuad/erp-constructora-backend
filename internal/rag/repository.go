package rag

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/pgvector/pgvector-go"
)

type ChunkWithDocument struct {
	DocumentID    string
	DocumentTitle string
	ChunkContent  string
	Score         float64
}

type Repository interface {
	SearchSimilarChunks(ctx context.Context, companyID string, projectID *string, queryEmbedding []float32, limit int) ([]ChunkWithDocument, error)
	SaveDocumentEmbedding(ctx context.Context, companyID, docID, versionID string, chunkIndex int, content string, vector []float32) error
	DeleteDocumentEmbeddings(ctx context.Context, companyID, docID string) error
}

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repository{db: db}
}

func (r *repository) SearchSimilarChunks(ctx context.Context, companyID string, projectID *string, queryEmbedding []float32, limit int) ([]ChunkWithDocument, error) {
	// Búsqueda vectorial con Pgvector (<=> distancia coseno)
	// Unimos con la tabla 'documents' para saber el título del archivo consultado
	query := `
		SELECT 
			e.document_id,
			d.title,
			e.chunk_content,
			(1 - (e.embedding <=> $1)) AS similarity
		FROM document_embeddings e
		INNER JOIN documents d ON d.id = e.document_id
		WHERE e.company_id = $2
	`

	args := []interface{}{pgvector.NewVector(queryEmbedding), companyID}
	argCount := 2

	// Si el usuario consulta desde dentro de un proyecto específico, filtramos por proyecto
	if projectID != nil && *projectID != "" {
		argCount++
		query += fmt.Sprintf(" AND d.project_id = $%d", argCount)
		args = append(args, *projectID)
	}

	argCount++
	query += fmt.Sprintf(" ORDER BY e.embedding <=> $1 ASC LIMIT $%d;", argCount)
	args = append(args, limit)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("error al realizar la búsqueda vectorial: %w", err)
	}
	defer rows.Close()

	var results []ChunkWithDocument
	for rows.Next() {
		var item ChunkWithDocument
		if err := rows.Scan(&item.DocumentID, &item.DocumentTitle, &item.ChunkContent, &item.Score); err != nil {
			return nil, err
		}
		results = append(results, item)
	}

	return results, nil
}
func (r *repository) SaveDocumentEmbedding(ctx context.Context, companyID, docID, versionID string, chunkIndex int, content string, vector []float32) error {
	query := `
		INSERT INTO document_embeddings (company_id, document_id, document_version_id, chunk_index, chunk_content, embedding)
		VALUES ($1, $2, $3, $4, $5, $6);
	`
	_, err := r.db.ExecContext(ctx, query, companyID, docID, versionID, chunkIndex, content, pgvector.NewVector(vector))
	return err
}

func (r *repository) DeleteDocumentEmbeddings(ctx context.Context, companyID, docID string) error {
	query := `DELETE FROM document_embeddings WHERE company_id = $1 AND document_id = $2`
	_, err := r.db.ExecContext(ctx, query, companyID, docID)
	return err
}
