package project

import (
	"context"
	"errors"
	"strings"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) UpdateProject(ctx context.Context, companyID string, id string, dto UpdateProjectDTO) (*Project, error) {
	if dto.Name != nil && *dto.Name == "" {
		return nil, errors.New("el nombre del proyecto no puede estar vacío")
	}
	if dto.Budget != nil && *dto.Budget <= 0 {
		return nil, errors.New("el presupuesto del proyecto debe ser mayor a cero")
	}
	if dto.StartDate != nil && dto.EndDate != nil && dto.EndDate.Before(*dto.StartDate) {
		return nil, errors.New("la fecha de fin no puede ser anterior a la fecha de inicio")
	}

	err := s.repo.Update(ctx, companyID, id, dto)
	if err != nil {
		return nil, err
	}

	return s.repo.GetByID(ctx, companyID, id)
}

func (s *Service) DeleteProject(ctx context.Context, companyID string, id string) error {
	err := s.repo.Delete(ctx, companyID, id)
	if err != nil && strings.Contains(err.Error(), "23503") {
		return errors.New("no se puede eliminar el proyecto porque tiene datos relacionados (presupuestos, gastos, órdenes de compra, etc.)")
	}
	return err
}

func (s *Service) CreateProject(ctx context.Context, companyID string, dto CreateProjectDTO) (*Project, error) {
	// Regla de Negocio 1: Validaciones de lógica financiera y de fechas
	if dto.Budget <= 0 {
		return nil, errors.New("el presupuesto del proyecto debe ser mayor a cero")
	}
	if dto.EndDate.Before(dto.StartDate) {
		return nil, errors.New("la fecha de fin no puede ser anterior a la fecha de inicio")
	}

	// Mapear DTO a la entidad del dominio
	project := &Project{
		CompanyID: companyID,
		Name:      dto.Name,
		ClientID:  dto.ClientID, // En un paso posterior validaremos si el cliente existe
		Location:  dto.Location,
		StartDate: dto.StartDate,
		EndDate:   dto.EndDate,
		Budget:    dto.Budget,
	}

	err := s.repo.Save(ctx, project)
	if err != nil {
		return nil, err
	}

	return project, nil
}

func (s *Service) ListProjects(ctx context.Context, companyID string) ([]Project, error) {
	return s.repo.GetAll(ctx, companyID)
}
