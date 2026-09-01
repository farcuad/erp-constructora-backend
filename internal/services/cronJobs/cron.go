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
		log.Printf("Advertencia: timezone America/Caracas no disponible, usando UTC-4: %v", err)
		loc = time.FixedZone("America/Caracas", -4*60*60)
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
			`<div style="text-align:center;padding:10px 0;">
				<div style="display:inline-block;background-color:#F99B2E;color:#000000;padding:10px 25px;border-radius:6px;font-weight:800;font-size:14px;letter-spacing:1px;margin-bottom:25px;">AVISO IMPORTANTE</div>
				<h1 style="color:#000000;font-size:24px;font-weight:700;margin:0 0 15px 0;">Tu suscripción expira en 3 días</h1>
				<p style="color:#333333;font-size:16px;line-height:1.6;margin:0 0 10px 0;">La empresa con ID <strong style="color:#F99B2E;">%s</strong> tiene una suscripción que vence el <strong>%s</strong>.</p>
				<p style="color:#666666;font-size:14px;line-height:1.6;">Por favor, renueva tu plan para evitar interrupciones en tu servicio.</p>
			</div>`,
			sub.CompanyID, expiryDate,
		)

		if err := mailService.SendEmail(adminEmail, subject, htmlBody); err != nil {
			log.Printf("[CRON] Error enviando email a %s: %v", adminEmail, err)
		}
	}
}