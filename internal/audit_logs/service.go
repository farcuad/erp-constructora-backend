package audit

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

func (s *Service) RegisterLog(ctx context.Context, log *AuditLog) error {
	if log.CompanyID == "" || log.UserID == "" || log.Action == "" || log.TableName == "" {
		return errors.New("los campos obligatorios de auditoría no pueden estar vacíos")
	}
	return s.repo.Insert(ctx, log)
}

func (s *Service) FetchLogs(ctx context.Context, companyID string) ([]AuditLog, error) {
	if companyID == "" {
		return nil, errors.New("el id de la empresa es requerido")
	}
	return s.repo.GetByCompany(ctx, companyID)
}
