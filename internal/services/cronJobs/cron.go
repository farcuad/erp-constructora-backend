package cronJobs

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/robfig/cron/v3"
	"erp-constructora/internal/subscriptions"
	"erp-constructora/internal/services"
	"erp-constructora/internal/users"
)

func StartSubscriptionExpiryCron(ctx context.Context, subService *subscriptions.Service, mailService *services.MailService, userRepo *users.Repository) {
	loc, err := time.LoadLocation("America/Caracas")
	if err != nil {
		log.Printf("Error cargando timezone America/Caracas: %v", err)
		return
	}

	c := cron.New(cron.WithLocation(loc))

	_, err = c.AddFunc("0 6 * * *", func() {
		runSubscriptionCheck(ctx, subService, mailService, userRepo)
	})
	if err != nil {
		log.Printf("Error configurando cron job: %v", err)
		return
	}

	c.Start()
	log.Println("Cron job de suscripciones iniciado (6:00 AM America/Caracas)")
}

func runSubscriptionCheck(ctx context.Context, subService *subscriptions.Service, mailService *services.MailService, userRepo *users.Repository) {
	subs, err := subService.GetSubscriptionsExpiringSoon(ctx, 3)
	if err != nil {
		log.Printf("[CRON] Error obteniendo suscripciones expiradas: %v", err)
		return
	}

	for _, sub := range subs {
		adminEmail, err := userRepo.GetAdminEmailByCompanyID(ctx, sub.CompanyID)
		if err != nil {
			log.Printf("[CRON] Error obteniendo admin email para company %s: %v", sub.CompanyID, err)
			continue
		}
		if adminEmail == "" {
			continue
		}

		var expiryDate string
		if sub.TrialEndDate != nil {
			expiryDate = sub.TrialEndDate.Format("2006-01-02")
		} else if sub.EndDate != nil {
			expiryDate = sub.EndDate.Format("2006-01-02")
		}

		subject := "Aviso: Tu suscripción expira en 3 días"
		htmlBody := fmt.Sprintf(
			`<h1>Tu suscripción está por vencer</h1><p>La empresa con ID <strong>%s</strong> tiene una suscripción que expira el <strong>%s</strong> (3 días).</p><p>Por favor, renueva tu plan para evitar interrupciones.</p>`,
			sub.CompanyID, expiryDate,
		)

		if err := mailService.SendEmail(adminEmail, subject, htmlBody); err != nil {
			log.Printf("[CRON] Error enviando email a %s: %v", adminEmail, err)
		}
	}
}