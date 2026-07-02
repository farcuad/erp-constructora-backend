package attendance

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

func (s *Service) SubmitAttendance(ctx context.Context, attendance *Attendance) error {
	if attendance.ProjectID == "" || attendance.Date == "" {
		return errors.New("el proyecto y la fecha son obligatorios")
	}
	if len(attendance.Logs) == 0 {
		return errors.New("la planilla debe incluir el registro de al menos un empleado")
	}

	// Validaciones de negocio básicas
	for _, log := range attendance.Logs {
		if log.EmployeeID == "" || log.Status == "" {
			return errors.New("el id del empleado y su estado de asistencia son requeridos")
		}
		if log.HoursWorked < 0 {
			return errors.New("las horas trabajadas no pueden ser negativas")
		}
	}

	return s.repo.SaveDailyAttendance(ctx, attendance)
}

func (s *Service) GetDailyReport(ctx context.Context, projectID string, date string) (*Attendance, error) {
	if projectID == "" || date == "" {
		return nil, errors.New("el proyecto y la fecha son requeridos")
	}
	return s.repo.GetAttendanceByProjectAndDate(ctx, projectID, date)
}
