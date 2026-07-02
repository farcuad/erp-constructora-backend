package equipement

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

func (s *Service) RegisterEquipment(ctx context.Context, e *Equipment) error {
	if e.Name == "" {
		return errors.New("el nombre de la maquinaria/equipo es mandatorio")
	}
	if e.OwnershipType == "" {
		e.OwnershipType = "Owned"
	}
	return s.repo.CreateEquipment(ctx, e)
}

func (s *Service) ListEquipment(ctx context.Context, companyID string) ([]Equipment, error) {
	return s.repo.GetEquipmentByCompany(ctx, companyID)
}

func (s *Service) AssignEquipment(ctx context.Context, a *EquipmentAssignment) error {
	if a.EquipmentID == "" || a.ProjectID == "" || a.StartDate == "" {
		return errors.New("id del equipo, proyecto y fecha de inicio son requeridos")
	}
	return s.repo.AssignToProject(ctx, a)
}

func (s *Service) RegisterMaintenance(ctx context.Context, m *MaintenanceRecord) error {
	if m.EquipmentID == "" || m.MaintenanceType == "" || m.MaintenanceDate == "" {
		return errors.New("equipo, tipo de mantenimiento y fecha son requeridos")
	}
	return s.repo.CreateMaintenance(ctx, m)
}
