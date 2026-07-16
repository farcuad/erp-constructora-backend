package schedule

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

func (s *Service) CreateTask(ctx context.Context, t *Task) error {
	if t.ProjectID == "" || t.Name == "" {
		return errors.New("proyecto y nombre de tarea son obligatorios")
	}
	return s.repo.CreateTask(ctx, t)
}

func (s *Service) GetProjectTasks(ctx context.Context, projectID string) ([]Task, error) {
	return s.repo.GetByProject(ctx, projectID)
}

func (s *Service) UpdateTask(ctx context.Context, t *Task) error {
	if t.Name == "" {
		return errors.New("el nombre de la tarea es obligatorio")
	}
	return s.repo.UpdateTask(ctx, t)
}

func (s *Service) DeleteTask(ctx context.Context, id string) error {
	return s.repo.DeleteTask(ctx, id)
}
