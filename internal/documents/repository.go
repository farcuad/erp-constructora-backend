package documents

import (
	"context"
	"database/sql"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateType(ctx context.Context, t *DocumentType) error {
	query := `INSERT INTO document_types (company_id, name, description) VALUES ($1, $2, $3) RETURNING id, created_at`
	return r.db.QueryRowContext(ctx, query, t.CompanyID, t.Name, t.Description).Scan(&t.ID, &t.CreatedAt)
}

func (r *Repository) CreateDocumentWithVersion(ctx context.Context, doc *Document, ver *DocumentVersion) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 1. Insertar el documento base
	queryDoc := `
		INSERT INTO documents (company_id, project_id, document_type_id, title, description, current_version, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at, updated_at`

	err = tx.QueryRowContext(ctx, queryDoc, doc.CompanyID, doc.ProjectID, doc.DocumentTypeID, doc.Title, doc.Description, 1, "ACTIVE").
		Scan(&doc.ID, &doc.CreatedAt, &doc.UpdatedAt)
	if err != nil {
		return err
	}

	// 2. Insertar la primera versión ligada al ID recién generado
	queryVer := `
		INSERT INTO document_versions (company_id, document_id, version_number, file_url, file_size, file_extension, change_log, user_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at`

	ver.CompanyID = doc.CompanyID
	ver.DocumentID = doc.ID
	ver.VersionNumber = 1

	err = tx.QueryRowContext(ctx, queryVer, ver.CompanyID, ver.DocumentID, ver.VersionNumber, ver.FileURL, ver.FileSize, ver.FileExtension, ver.ChangeLog, ver.UserID).
		Scan(&ver.ID, &ver.CreatedAt)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *Repository) AddNewVersion(ctx context.Context, ver *DocumentVersion) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 1. Obtener la versión actual e incrementarla
	var currentVer int
	queryCheck := `SELECT current_version FROM documents WHERE id = $1 AND company_id = $2 FOR UPDATE`
	err = tx.QueryRowContext(ctx, queryCheck, ver.DocumentID, ver.CompanyID).Scan(&currentVer)
	if err != nil {
		return err
	}

	nextVer := currentVer + 1
	ver.VersionNumber = nextVer

	// 2. Insertar nueva versión
	queryVer := `
		INSERT INTO document_versions (company_id, document_id, version_number, file_url, file_size, file_extension, change_log, user_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at`

	err = tx.QueryRowContext(ctx, queryVer, ver.CompanyID, ver.DocumentID, ver.VersionNumber, ver.FileURL, ver.FileSize, ver.FileExtension, ver.ChangeLog, ver.UserID).
		Scan(&ver.ID, &ver.CreatedAt)
	if err != nil {
		return err
	}

	// 3. Actualizar la cabecera del documento con el nuevo número de versión
	queryUpdateDoc := `UPDATE documents SET current_version = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2 AND company_id = $3`
	_, err = tx.ExecContext(ctx, queryUpdateDoc, nextVer, ver.DocumentID, ver.CompanyID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *Repository) UpdateType(ctx context.Context, companyID, id string, req UpdateDocumentTypeRequest) error {
	query := `
		UPDATE document_types
		SET name = COALESCE($1, name),
		    description = COALESCE($2, description)
		WHERE company_id = $3 AND id = $4`

	var name, desc interface{}
	if req.Name != nil {
		name = *req.Name
	} else {
		name = nil
	}
	if req.Description != nil {
		desc = *req.Description
	} else {
		desc = nil
	}

	_, err := r.db.ExecContext(ctx, query, name, desc, companyID, id)
	return err
}

func (r *Repository) DeleteType(ctx context.Context, companyID, id string) error {
	query := `DELETE FROM document_types WHERE company_id = $1 AND id = $2`
	_, err := r.db.ExecContext(ctx, query, companyID, id)
	return err
}

func (r *Repository) UpdateDocument(ctx context.Context, companyID, id string, req UpdateDocumentRequest) error {
	query := `
		UPDATE documents
		SET title = COALESCE($1, title),
		    description = COALESCE($2, description),
		    document_type_id = COALESCE($3, document_type_id),
		    updated_at = CURRENT_TIMESTAMP
		WHERE company_id = $4 AND id = $5`

	var title, desc, dtID interface{}
	if req.Title != nil {
		title = *req.Title
	} else {
		title = nil
	}
	if req.Description != nil {
		desc = *req.Description
	} else {
		desc = nil
	}
	if req.DocumentTypeID != nil {
		dtID = *req.DocumentTypeID
	} else {
		dtID = nil
	}

	_, err := r.db.ExecContext(ctx, query, title, desc, dtID, companyID, id)
	return err
}

func (r *Repository) DeleteDocument(ctx context.Context, companyID, id string) error {
	query := `DELETE FROM documents WHERE company_id = $1 AND id = $2`
	_, err := r.db.ExecContext(ctx, query, companyID, id)
	return err
}
