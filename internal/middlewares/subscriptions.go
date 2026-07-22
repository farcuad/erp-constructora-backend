package middlewares

import (
	"context"
	"errors"
	"net/http"
)

type SubscriptionKeyType string

const SubscriptionKey SubscriptionKeyType = "subscription"

type SubscriptionInfo struct {
	ID          string `json:"id"`
	CompanyID   string `json:"company_id"`
	Status      string `json:"status"`
	MaxProjects int    `json:"max_projects"`
	MaxUsers    int    `json:"max_users"`
}

type SubscriptionService interface {
	IsSubscriptionActive(ctx context.Context, companyID string) (bool, error)
	GetSubscriptionInfo(ctx context.Context, companyID string) (*SubscriptionInfo, error)
	CanCreateProject(ctx context.Context, companyID string) (bool, error)
}

func RequireActiveSubscription(svc SubscriptionService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			companyID, ok := GetCompanyIDFromContext(r.Context())
			if !ok || companyID == "" {
				http.Error(w, "No autorizado", http.StatusUnauthorized)
				return
			}

			active, err := svc.IsSubscriptionActive(r.Context(), companyID)
			if err != nil || !active {
				http.Error(w, "Suscripción inactiva o expirada", http.StatusPaymentRequired)
				return
			}

			info, err := svc.GetSubscriptionInfo(r.Context(), companyID)
			if err == nil && info != nil {
				ctx := context.WithValue(r.Context(), SubscriptionKey, info)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func GetSubscriptionFromContext(ctx context.Context) (*SubscriptionInfo, bool) {
	sub, ok := ctx.Value(SubscriptionKey).(*SubscriptionInfo)
	return sub, ok
}

var ErrProjectLimitExceeded = errors.New("límite de proyectos alcanzado para tu plan")
var ErrUserLimitExceeded = errors.New("límite de usuarios alcanzado para tu plan")
