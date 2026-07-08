package users

import (
	"context"

	"erp-constructora/internal/middlewares"
)

func GetCompanyIDFromContext(ctx context.Context) (string, bool) {
	return middlewares.GetCompanyIDFromContext(ctx)
}

func GetUserIDFromContext(ctx context.Context) (string, bool) {
	return middlewares.GetUserIDFromContext(ctx)
}
