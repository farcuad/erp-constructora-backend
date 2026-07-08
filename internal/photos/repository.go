package photos

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

func (r *Repository) Save(ctx context.Context, p *ProjectPhoto) error {
	query := `
		INSERT INTO project_photos (company_id, project_id, task_id, daily_report_id, user_id, photo_url, description, latitude, longitude)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, created_at`

	return r.db.QueryRowContext(ctx, query,
		p.CompanyID,
		p.ProjectID,
		p.TaskID,
		p.DailyReportID,
		p.UserID,
		p.PhotoURL,
		p.Description,
		p.Latitude,
		p.Longitude,
	).Scan(&p.ID, &p.CreatedAt)
}

func (r *Repository) GetByProject(ctx context.Context, companyID, projectID string) ([]ProjectPhoto, error) {
	query := `
		SELECT id, company_id, project_id, task_id, daily_report_id, user_id, photo_url, description, latitude, longitude, created_at
		FROM project_photos
		WHERE company_id = $1 AND project_id = $2
		ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, companyID, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var photos []ProjectPhoto
	for rows.Next() {
		var p ProjectPhoto
		err := rows.Scan(
			&p.ID, &p.CompanyID, &p.ProjectID, &p.TaskID, &p.DailyReportID,
			&p.UserID, &p.PhotoURL, &p.Description, &p.Latitude, &p.Longitude, &p.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		photos = append(photos, p)
	}
	return photos, nil
}

func (r *Repository) Update(ctx context.Context, companyID, id string, req UpdatePhotoRequest) error {
	query := `
		UPDATE project_photos
		SET description = COALESCE($1, description),
		    latitude = COALESCE($2, latitude),
		    longitude = COALESCE($3, longitude)
		WHERE company_id = $4 AND id = $5`

	var desc interface{}
	if req.Description != nil {
		desc = *req.Description
	} else {
		desc = nil
	}

	_, err := r.db.ExecContext(ctx, query, desc, req.Latitude, req.Longitude, companyID, id)
	return err
}

func (r *Repository) Delete(ctx context.Context, companyID, id string) error {
	query := `DELETE FROM project_photos WHERE company_id = $1 AND id = $2`
	_, err := r.db.ExecContext(ctx, query, companyID, id)
	return err
}
