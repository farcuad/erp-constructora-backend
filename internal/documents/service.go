package documents

import (
	"context"
	"errors"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateDocumentType(ctx context.Context, t *DocumentType) error {
	if t.Name == "" {
		return errors.New("el nombre del tipo de documento es requerido")
	}
	return s.repo.CreateType(ctx, t)
}

func (s *Service) UploadInitialDocument(ctx context.Context, doc *Document, ver *DocumentVersion) error {
	if doc.ProjectID == "" || doc.DocumentTypeID == "" || doc.Title == "" || ver.FileURL == "" {
		return errors.New("faltan campos obligatorios para registrar el documento")
	}
	return s.repo.CreateDocumentWithVersion(ctx, doc, ver)
}

func (s *Service) UploadNewVersion(ctx context.Context, ver *DocumentVersion) error {
	if ver.DocumentID == "" || ver.FileURL == "" || ver.UserID == "" {
		return errors.New("el id del documento y la url del archivo son obligatorios para una nueva versión")
	}
	return s.repo.AddNewVersion(ctx, ver)
}
