package notifications

import (
	"context"
	"database/sql"
	"errors"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) DispatchNotification(ctx context.Context, n *Notification, targetUserIDs []string) error {
	if n.Title == "" || n.Message == "" {
		return errors.New("el título y el cuerpo de la notificación son obligatorios")
	}
	if len(targetUserIDs) == 0 {
		return errors.New("debe especificar al menos un usuario destino")
	}

	// Ejecutamos ambas inserciones bajo una única transacción de Base de Datos
	return s.repo.ExecInTx(ctx, func(tx *sql.Tx) error {
		if err := s.repo.CreateTx(ctx, tx, n); err != nil {
			return err
		}

		if err := s.repo.AssignToUsersBulk(ctx, tx, n.CompanyID, n.ID, targetUserIDs); err != nil {
			return err
		}
		return nil
	})
}

func (s *Service) FetchMyNotifications(ctx context.Context, companyID, userID string) ([]Notification, error) {
	if companyID == "" || userID == "" {
		return nil, errors.New("identificadores no válidos")
	}
	return s.repo.GetUserNotifications(ctx, companyID, userID)
}

func (s *Service) DeleteNotification(ctx context.Context, companyID, id string) error {
	if companyID == "" || id == "" {
		return errors.New("el id de la empresa y de la notificación son requeridos")
	}
	return s.repo.Delete(ctx, companyID, id)
}

func (s *Service) ReadNotification(ctx context.Context, companyID, notificationID, userID string) error {
	if notificationID == "" {
		return errors.New("el id de la notificación es requerido")
	}
	return s.repo.MarkAsRead(ctx, companyID, notificationID, userID)
}
