package fcm

import (
	"context"
	"log"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"google.golang.org/api/option"
)

type FCMClient struct {
	client *messaging.Client
}

func NewFCMClient(credentialsPath string) (*FCMClient, error) {
	ctx := context.Background()
	opt := option.WithCredentialsFile(credentialsPath)

	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		return nil, err
	}

	messagingClient, err := app.Messaging(ctx)
	if err != nil {
		return nil, err
	}

	return &FCMClient{client: messagingClient}, nil
}

// NewFCMClientFromJSON inicializa el cliente con el contenido crudo del JSON de la
// cuenta de servicio (ideal para deploys donde el archivo no está en el repo,
// ej: variable de entorno FIREBASE_CREDENTIALS_JSON en un VPS / contenedor).
func NewFCMClientFromJSON(credentialsJSON []byte) (*FCMClient, error) {
	ctx := context.Background()
	opt := option.WithCredentialsJSON(credentialsJSON)

	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		return nil, err
	}

	messagingClient, err := app.Messaging(ctx)
	if err != nil {
		return nil, err
	}

	return &FCMClient{client: messagingClient}, nil
}

// PushNotification encapsula el envío individual o múltiple
func (f *FCMClient) SendPush(ctx context.Context, tokens []string, title, body string, data map[string]string) {
	if len(tokens) == 0 {
		return
	}

	// Si solo hay un token, usaremos Send; para varios, SendMulticast
	message := &messaging.MulticastMessage{
		Tokens: tokens,
		Notification: &messaging.Notification{
			Title: title,
			Body:  body,
		},
		Data: data,
		Android: &messaging.AndroidConfig{
			Priority: "high",
			Notification: &messaging.AndroidNotification{
				Sound: "default",
			},
		},
	}

	response, err := f.client.SendEachForMulticast(ctx, message)
	if err != nil {
		log.Printf("Error al enviar FCM multicast: %v", err)
		return
	}

	log.Printf("Notificaciones enviadas con éxito: %d, fallidas: %d", response.SuccessCount, response.FailureCount)
}
