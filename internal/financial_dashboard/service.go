package financialdashboard

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

func (s *Service) GetDashboardKPIs(ctx context.Context, companyID, projectID string) (*ProjectKPIs, error) {
	if companyID == "" || projectID == "" {
		return nil, errors.New("company_id y project_id son requeridos")
	}
	return s.repo.GetProjectFinancialSummary(ctx, companyID, projectID)
}
