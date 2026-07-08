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
	if t.ProjectID == "" || t.Name == "" || t.StartDate == "" || t.EndDate == "" {
		return errors.New("proyecto, nombre de tarea, fecha de inicio y fin son mandatorios")
	}
	return s.repo.CreateTask(ctx, t)
}

func (s *Service) LinkTasks(ctx context.Context, d *TaskDependency) error {
	if d.TaskID == "" || d.DependsOnUUID == "" {
		return errors.New("se requiere el id de la tarea sucesora y de la predecesora")
	}
	if d.TaskID == d.DependsOnUUID {
		return errors.New("una tarea no puede depender de sí misma")
	}
	if d.DependencyType == "" {
		d.DependencyType = "FS" // Finish to Start por defecto
	}
	return s.repo.AddDependency(ctx, d)
}

func (s *Service) CreateMilestone(ctx context.Context, m *Milestone) error {
	if m.ProjectID == "" || m.Name == "" || m.DueDate == "" {
		return errors.New("proyecto, nombre e hito temporal son requeridos")
	}
	return s.repo.CreateMilestone(ctx, m)
}

func (s *Service) GetProjectTasks(ctx context.Context, projectID string) ([]Task, error) {
	return s.repo.GetScheduleByProject(ctx, projectID)
}

func (s *Service) UpdateTask(ctx context.Context, t *Task) error {
	if t.Name == "" || t.StartDate == "" || t.EndDate == "" {
		return errors.New("nombre de tarea, fecha de inicio y fin son mandatorios")
	}
	return s.repo.UpdateTask(ctx, t)
}

func (s *Service) DeleteTask(ctx context.Context, id string) error {
	return s.repo.DeleteTask(ctx, id)
}

func (s *Service) UpdateMilestone(ctx context.Context, m *Milestone) error {
	if m.Name == "" || m.DueDate == "" {
		return errors.New("nombre y fecha del hito son requeridos")
	}
	return s.repo.UpdateMilestone(ctx, m)
}

func (s *Service) DeleteMilestone(ctx context.Context, id string) error {
	return s.repo.DeleteMilestone(ctx, id)
}
