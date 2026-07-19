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

func (s *Service) UpdateDocumentType(ctx context.Context, companyID, id string, req UpdateDocumentTypeRequest) error {
	if companyID == "" || id == "" {
		return errors.New("el id de la empresa y del tipo de documento son requeridos")
	}
	return s.repo.UpdateType(ctx, companyID, id, req)
}

func (s *Service) DeleteDocumentType(ctx context.Context, companyID, id string) error {
	if companyID == "" || id == "" {
		return errors.New("el id de la empresa y del tipo de documento son requeridos")
	}
	return s.repo.DeleteType(ctx, companyID, id)
}

func (s *Service) UpdateDocument(ctx context.Context, companyID, id string, req UpdateDocumentRequest) error {
	if companyID == "" || id == "" {
		return errors.New("el id de la empresa y del documento son requeridos")
	}
	return s.repo.UpdateDocument(ctx, companyID, id, req)
}

func (s *Service) DeleteDocument(ctx context.Context, companyID, id string) error {
	if companyID == "" || id == "" {
		return errors.New("el id de la empresa y del documento son requeridos")
	}
	return s.repo.DeleteDocument(ctx, companyID, id)
}

func (s *Service) GetDocumentTypes(ctx context.Context, companyID string) ([]DocumentType, error) {
	if companyID == "" {
		return nil, errors.New("empresa requerida")
	}
	return s.repo.GetTypes(ctx, companyID)
}

func (s *Service) GetProjectDocuments(ctx context.Context, companyID, projectID string) ([]Document, error) {
	if companyID == "" || projectID == "" {
		return nil, errors.New("empresa y proyecto son requeridos")
	}
	return s.repo.GetByProject(ctx, companyID, projectID)
}

func (s *Service) GetDocumentByID(ctx context.Context, companyID, id string) (*Document, error) {
	if companyID == "" || id == "" {
		return nil, errors.New("empresa e id de documento son requeridos")
	}
	return s.repo.GetByID(ctx, companyID, id)
}

func (s *Service) GetDocumentVersions(ctx context.Context, companyID, documentID string) ([]DocumentVersion, error) {
	if companyID == "" || documentID == "" {
		return nil, errors.New("empresa e id de documento son requeridos")
	}
	return s.repo.GetVersions(ctx, companyID, documentID)
}
