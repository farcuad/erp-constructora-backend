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

func (r *Repository) GetTypes(ctx context.Context, companyID string) ([]DocumentType, error) {
	query := `SELECT id, company_id, name, COALESCE(description, ''), created_at FROM document_types WHERE company_id = $1 ORDER BY name ASC`

	rows, err := r.db.QueryContext(ctx, query, companyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var types []DocumentType
	for rows.Next() {
		var t DocumentType
		if err := rows.Scan(&t.ID, &t.CompanyID, &t.Name, &t.Description, &t.CreatedAt); err != nil {
			return nil, err
		}
		types = append(types, t)
	}
	return types, nil
}

func (r *Repository) GetByProject(ctx context.Context, companyID, projectID string) ([]Document, error) {
	query := `SELECT id, company_id, project_id, document_type_id, title, COALESCE(description, ''), current_version, status, created_at, updated_at FROM documents WHERE company_id = $1 AND project_id = $2 ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, companyID, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var docs []Document
	for rows.Next() {
		var d Document
		if err := rows.Scan(&d.ID, &d.CompanyID, &d.ProjectID, &d.DocumentTypeID, &d.Title, &d.Description, &d.CurrentVersion, &d.Status, &d.CreatedAt, &d.UpdatedAt); err != nil {
			return nil, err
		}
		docs = append(docs, d)
	}
	return docs, nil
}

func (r *Repository) GetByID(ctx context.Context, companyID, id string) (*Document, error) {
	query := `SELECT id, company_id, project_id, document_type_id, title, COALESCE(description, ''), current_version, status, created_at, updated_at FROM documents WHERE company_id = $1 AND id = $2`

	var d Document
	if err := r.db.QueryRowContext(ctx, query, companyID, id).Scan(&d.ID, &d.CompanyID, &d.ProjectID, &d.DocumentTypeID, &d.Title, &d.Description, &d.CurrentVersion, &d.Status, &d.CreatedAt, &d.UpdatedAt); err != nil {
		return nil, err
	}

	versions, err := r.GetVersions(ctx, companyID, d.ID)
	if err != nil {
		return nil, err
	}
	d.Versions = versions

	return &d, nil
}

func (r *Repository) GetVersions(ctx context.Context, companyID, documentID string) ([]DocumentVersion, error) {
	query := `SELECT id, company_id, document_id, version_number, file_url, COALESCE(file_size, 0), COALESCE(file_extension, ''), COALESCE(change_log, ''), user_id, created_at FROM document_versions WHERE company_id = $1 AND document_id = $2 ORDER BY version_number DESC`

	rows, err := r.db.QueryContext(ctx, query, companyID, documentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var versions []DocumentVersion
	for rows.Next() {
		var v DocumentVersion
		if err := rows.Scan(&v.ID, &v.CompanyID, &v.DocumentID, &v.VersionNumber, &v.FileURL, &v.FileSize, &v.FileExtension, &v.ChangeLog, &v.UserID, &v.CreatedAt); err != nil {
			return nil, err
		}
		versions = append(versions, v)
	}
	return versions, nil
}
