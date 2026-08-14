package documents

import (
	"context"
	"errors"
	"log"
	"time"
)

// DocumentIndexer indexa semánticamente (RAG) una versión de documento.
type DocumentIndexer interface {
	IndexDocument(ctx context.Context, companyID, docID, versionID, fileURL, fileExtension string) error
	DeleteDocumentEmbeddings(ctx context.Context, companyID, docID string) error
}

type Service struct {
	repo    *Repository
	indexer DocumentIndexer
}

func NewService(repo *Repository, indexer DocumentIndexer) *Service {
	return &Service{repo: repo, indexer: indexer}
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
	if err := s.repo.CreateDocumentWithVersion(ctx, doc, ver); err != nil {
		return err
	}
	s.indexInBackground(doc.CompanyID, doc.ID, ver.ID, ver.FileURL, ver.FileExtension)
	return nil
}

func (s *Service) UploadNewVersion(ctx context.Context, ver *DocumentVersion) error {
	if ver.DocumentID == "" || ver.FileURL == "" || ver.UserID == "" {
		return errors.New("el id del documento y la url del archivo son obligatorios para una nueva versión")
	}
	if err := s.repo.AddNewVersion(ctx, ver); err != nil {
		return err
	}
	s.indexInBackground(ver.CompanyID, ver.DocumentID, ver.ID, ver.FileURL, ver.FileExtension)
	return nil
}

// indexInBackground delega el procesamiento RAG (parseo, troceo y embeddings)
// a un goroutine para no bloquear la respuesta HTTP del POST.
func (s *Service) indexInBackground(companyID, docID, versionID, fileURL, fileExtension string) {
	if s.indexer == nil {
		return
	}
	go func() {
		idxCtx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()

		// Reemplaza embeddings de versiones anteriores del mismo documento
		if err := s.indexer.DeleteDocumentEmbeddings(idxCtx, companyID, docID); err != nil {
			log.Printf("[RAG] error limpiando embeddings del documento %s: %v", docID, err)
		}
		if err := s.indexer.IndexDocument(idxCtx, companyID, docID, versionID, fileURL, fileExtension); err != nil {
			log.Printf("[RAG] error indexando documento %s (versión %s): %v", docID, versionID, err)
		}
	}()
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
	if err := s.repo.DeleteDocument(ctx, companyID, id); err != nil {
		return err
	}
	if s.indexer != nil {
		// La FK document_embeddings.document_id es ON DELETE CASCADE, pero
		// limpiamos también por si el borrado dejara huérfanos.
		if err := s.indexer.DeleteDocumentEmbeddings(ctx, companyID, id); err != nil {
			log.Printf("[RAG] error limpiando embeddings del documento %s: %v", id, err)
		}
	}
	return nil
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
