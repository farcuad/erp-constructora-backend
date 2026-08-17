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

func (s *Service) GetDashboardKPIs(ctx context.Context, companyID, projectID string) (*AdvancedDashboardKPIs, error) {
	if companyID == "" || projectID == "" {
		return nil, errors.New("company_id y project_id son requeridos")
	}
	// 1. Obtener KPIs principales
	kpis, err := s.repo.GetProjectFinancialSummary(ctx, companyID, projectID)
	if err != nil {
		return nil, err
	}

	// 2. Obtener datos mensuales
	trends, err := s.repo.GetMonthlyTrends(ctx, companyID, projectID)
	if err != nil {
		return nil, err
	}

	// 3. Empaquetar todo
	return &AdvancedDashboardKPIs{
		ProjectKPIs:   *kpis,
		MonthlyTrends: trends,
	}, nil
}
