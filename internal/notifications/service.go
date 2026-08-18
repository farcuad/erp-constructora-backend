package notifications

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"erp-constructora/internal/middlewares"
)

type Service struct {
	repo *Repository
	hub  *WSHub
}

// Notifier describe el método que cualquier módulo puede usar para emitir notificaciones a toda la empresa
type Notifier interface {
	NotifyFromContext(ctx context.Context, req CreateNotificationRequest) error
}

func NewService(repo *Repository, hub *WSHub) *Service {
	return &Service{repo: repo, hub: hub}
}

// NotifyFromContext emite una notificación a toda la empresa. Lee company_id y el actor
// (quien ejecutó la acción) directamente del contexto JWT, y el actor queda excluido.
// Es el método que invocan los handlers de los demás módulos tras una mutación exitosa.
func (s *Service) NotifyFromContext(ctx context.Context, req CreateNotificationRequest) error {
	companyID, ok := middlewares.GetCompanyIDFromContext(ctx)
	if !ok {
		return errors.New("no se encontró la empresa en el contexto")
	}

	actorID, _ := middlewares.GetUserIDFromContext(ctx)

	_, err := s.NotifyAll(ctx, companyID, actorID, req)
	return err
}

// NotifyAll emite una notificación a todos los usuarios activos de la empresa, excluyendo al actor
// que ejecutó la acción (companyID y actorID salen del contexto JWT).
func (s *Service) NotifyAll(ctx context.Context, companyID, actorID string, req CreateNotificationRequest) (*Notification, error) {
	userIDs, err := s.repo.GetUserIDsByCompany(ctx, companyID)
	if err != nil {
		return nil, err
	}

	targets := make([]string, 0, len(userIDs))
	for _, id := range userIDs {
		if actorID != "" && id == actorID {
			continue
		}
		targets = append(targets, id)
	}

	if len(targets) == 0 {
		return nil, nil
	}

	req.TargetUsers = targets
	return s.DispatchNotification(ctx, req, companyID)
}

func (s *Service) DispatchNotification(ctx context.Context, req CreateNotificationRequest, companyID string) (*Notification, error) {
	if req.Title == "" || req.Message == "" {
		return nil, errors.New("el título y el mensaje son obligatorios")
	}
	if len(req.TargetUsers) == 0 {
		return nil, errors.New("debe especificar al menos un usuario destino")
	}

	// Asignar valores predeterminados si vienen vacíos
	if req.EntityType == "" {
		req.EntityType = "SYSTEM"
	}
	if req.Type == "" {
		req.Type = "GENERAL_ALERT"
	}
	if req.Priority == "" {
		req.Priority = PriorityMedium
	}

	// Serializar la metadata map -> json.RawMessage
	var metaBytes json.RawMessage
	if req.Metadata != nil {
		b, err := json.Marshal(req.Metadata)
		if err == nil {
			metaBytes = b
		}
	}

	notification := &Notification{
		CompanyID:  companyID,
		ProjectID:  req.ProjectID,
		EntityType: req.EntityType,
		EntityID:   req.EntityID,
		Type:       req.Type,
		Priority:   req.Priority,
		Title:      req.Title,
		Message:    req.Message,
		LinkToUI:   req.LinkToUI,
		Metadata:   metaBytes,
	}

	// 1. Guardar en Base de Datos en Transacción
	err := s.repo.ExecInTx(ctx, func(tx *sql.Tx) error {
		if err := s.repo.CreateTx(ctx, tx, notification); err != nil {
			return err
		}
		return s.repo.AssignToUsersBulk(ctx, tx, companyID, notification.ID, req.TargetUsers)
	})

	if err != nil {
		return nil, err
	}

	// 2. Transmisión vía WebSockets en tiempo real a los usuarios destino conectados
	s.hub.SendToUsers(companyID, req.TargetUsers, *notification)

	return notification, nil
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
