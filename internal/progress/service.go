package progress

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

func (s *Service) SaveDailyProgress(ctx context.Context, report *DailyReport) error {
	// Aquí podrías iniciar una transacción (tx) si deseas asegurar consistencia absoluta
	err := s.repo.CreateReport(ctx, report)
	if err != nil {
		return err
	}

	for i := range report.ProgressEntries {
		report.ProgressEntries[i].DailyReportID = report.ID
		report.ProgressEntries[i].CompanyID = report.CompanyID
		report.ProgressEntries[i].ProjectID = report.ProjectID

		err = s.repo.CreateProgressEntry(ctx, &report.ProgressEntries[i])
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *Service) GetDailyReport(ctx context.Context, companyID, projectID, date string) (*DailyReport, error) {
	if projectID == "" || date == "" {
		return nil, errors.New("el proyecto y la fecha son requeridos")
	}
	if companyID == "" {
		return nil, errors.New("el id de la empresa es requerido")
	}

	// Llamada limpia pasando los parámetros correspondientes
	return s.repo.GetReportWithProgress(ctx, companyID, projectID, date)
}

func (s *Service) UpdateDailyReport(ctx context.Context, companyID, id string, req UpdateDailyReportRequest) error {
	if companyID == "" || id == "" {
		return errors.New("el id de la empresa y del reporte son requeridos")
	}
	return s.repo.UpdateReport(ctx, companyID, id, req)
}

func (s *Service) DeleteDailyReport(ctx context.Context, companyID, id string) error {
	if companyID == "" || id == "" {
		return errors.New("el id de la empresa y del reporte son requeridos")
	}
	return s.repo.DeleteReport(ctx, companyID, id)
}
