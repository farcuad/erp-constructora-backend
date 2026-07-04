package photos

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

func (s *Service) RegisterPhoto(ctx context.Context, photo *ProjectPhoto) error {
	if photo.ProjectID == "" || photo.PhotoURL == "" {
		return errors.New("el id del proyecto y la url de la foto son campos obligatorios")
	}
	return s.repo.Save(ctx, photo)
}

func (s *Service) GetProjectGallery(ctx context.Context, companyID, projectID string) ([]ProjectPhoto, error) {
	if companyID == "" || projectID == "" {
		return nil, errors.New("el id de empresa y de proyecto son obligatorios")
	}
	return s.repo.GetByProject(ctx, companyID, projectID)
}
