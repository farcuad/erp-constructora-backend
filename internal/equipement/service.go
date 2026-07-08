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

func (s *Service) UpdateEquipment(ctx context.Context, id, companyID string, req *UpdateEquipmentRequest) (*Equipment, error) {
	e, err := s.repo.GetEquipmentByID(ctx, id, companyID)
	if err != nil {
		return nil, err
	}
	if req.Name != nil {
		e.Name = *req.Name
	}
	if req.TypeID != nil {
		e.TypeID = *req.TypeID
	}
	if req.PlateNumber != nil {
		e.PlateNumber = *req.PlateNumber
	}
	if req.Model != nil {
		e.Model = *req.Model
	}
	if req.Brand != nil {
		e.Brand = *req.Brand
	}
	if req.Status != nil {
		e.Status = *req.Status
	}
	if req.OwnershipType != nil {
		e.OwnershipType = *req.OwnershipType
	}
	if err := s.repo.UpdateEquipment(ctx, e); err != nil {
		return nil, err
	}
	return e, nil
}

func (s *Service) DeleteEquipment(ctx context.Context, id, companyID string) error {
	return s.repo.DeleteEquipment(ctx, id, companyID)
}

func (s *Service) UpdateEquipmentType(ctx context.Context, id, companyID string, req *UpdateEquipmentTypeRequest) (*EquipmentType, error) {
	et, err := s.repo.GetEquipmentTypeByID(ctx, id, companyID)
	if err != nil {
		return nil, err
	}
	if req.Name != nil {
		et.Name = *req.Name
	}
	if err := s.repo.UpdateEquipmentType(ctx, et); err != nil {
		return nil, err
	}
	return et, nil
}

func (s *Service) DeleteEquipmentType(ctx context.Context, id, companyID string) error {
	return s.repo.DeleteEquipmentType(ctx, id, companyID)
}
