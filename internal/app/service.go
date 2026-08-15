package app

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

// PublishRelease valida y guarda una nueva versión compilada de la app.
func (s *Service) PublishRelease(ctx context.Context, req *CreateReleaseRequest) (*Release, error) {
	if req.Version == "" {
		return nil, errors.New("la versión es obligatoria")
	}
	if req.AppURL == "" {
		return nil, errors.New("la url del archivo de la app es obligatoria")
	}
	return s.repo.Create(ctx, req)
}

// GetLatestRelease devuelve la subida más reciente de la app (puede ser nil).
func (s *Service) GetLatestRelease(ctx context.Context) (*Release, error) {
	return s.repo.GetLatest(ctx)
}
