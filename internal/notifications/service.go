package notifications

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log"

	"erp-constructora/internal/middlewares"
)

type Service struct {
	repo *Repository
	hub  *WSHub
	push PushSender
}

// PushSender envía push notifications vía FCM (implementado por pkg/fcm). Puede ser nil.
type PushSender interface {
	SendPush(ctx context.Context, tokens []string, title, body string, data map[string]string)
}

// Notifier describe el método que cualquier módulo puede usar para emitir notificaciones a toda la empresa
type Notifier interface {
	NotifyFromContext(ctx context.Context, req CreateNotificationRequest) error
}

func NewService(repo *Repository, hub *WSHub, push PushSender) *Service {
	return &Service{repo: repo, hub: hub, push: push}
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

	// 3. Push notification vía FCM para los que NO tienen la app abierta (asíncrono, no bloquea)
	if s.push != nil {
		tokens, err := s.repo.GetPushTokensByUsers(ctx, req.TargetUsers)
		if err != nil {
			log.Printf("[PUSH ERROR] no se pudieron obtener los tokens FCM: %v", err)
		} else if len(tokens) > 0 {
			data := map[string]string{
				"notification_id": notification.ID,
				"entity_type":     req.EntityType,
			}
			if req.EntityID != nil {
				data["entity_id"] = *req.EntityID
			}
			if req.ProjectID != nil {
				data["project_id"] = *req.ProjectID
			}
			if req.LinkToUI != nil {
				data["link_to_ui"] = *req.LinkToUI
			}
			// Contexto propio: el ctx del request muere al terminar el handler y cancelaría el envío
			go s.push.SendPush(context.Background(), tokens, req.Title, req.Message, data)
		}
	}

	return notification, nil
}

// RegisterPushToken guarda el token FCM que reporta la app móvil al iniciar sesión
func (s *Service) RegisterPushToken(ctx context.Context, userID, token, platform string) error {
	if userID == "" || token == "" {
		return errors.New("el token es obligatorio")
	}
	if platform != "android" && platform != "ios" {
		return errors.New("la plataforma debe ser 'android' o 'ios'")
	}
	return s.repo.SavePushToken(ctx, userID, token, platform)
}

// UnregisterPushToken elimina el token FCM (logout del dispositivo)
func (s *Service) UnregisterPushToken(ctx context.Context, token string) error {
	if token == "" {
		return errors.New("el token es obligatorio")
	}
	return s.repo.DeletePushToken(ctx, token)
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
