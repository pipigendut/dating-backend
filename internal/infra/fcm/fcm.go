package fcm

import (
	"context"
	"fmt"
	"log"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"google.golang.org/api/option"
)

type Client struct {
	app *firebase.App
	msg *messaging.Client
}

func NewClient(serviceAccountPath string) (*Client, error) {
	if serviceAccountPath == "" {
		return nil, fmt.Errorf("FIREBAR_SERVICE_ACCOUNT_PATH is required for FCM")
	}

	opt := option.WithCredentialsFile(serviceAccountPath)
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		return nil, fmt.Errorf("error initializing firebase app: %v", err)
	}

	msg, err := app.Messaging(context.Background())
	if err != nil {
		return nil, fmt.Errorf("error getting firebase messaging client: %v", err)
	}

	return &Client{
		app: app,
		msg: msg,
	}, nil
}

func (c *Client) SendToToken(ctx context.Context, token string, title, body string, data map[string]string) error {
	message := &messaging.Message{
		Token: token,
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
		APNS: &messaging.APNSConfig{
			Payload: &messaging.APNSPayload{
				Aps: &messaging.Aps{
					Sound: "default",
				},
			},
		},
	}

	_, err := c.msg.Send(ctx, message)
	if err != nil {
		log.Printf("[FCM] Error sending message to token %s: %v", token, err)
		return err
	}

	return nil
}

func (c *Client) SendMulticast(ctx context.Context, tokens []string, title, body string, data map[string]string) error {
	if len(tokens) == 0 {
		return nil
	}

	message := &messaging.MulticastMessage{
		Tokens: tokens,
		Notification: &messaging.Notification{
			Title: title,
			Body:  body,
		},
		Data: data,
		Android: &messaging.AndroidConfig{
			Priority: "high",
		},
		APNS: &messaging.APNSConfig{
			Payload: &messaging.APNSPayload{
				Aps: &messaging.Aps{
					Sound: "default",
				},
			},
		},
	}

	_, err := c.msg.SendMulticast(ctx, message)
	if err != nil {
		log.Printf("[FCM] Error sending multicast message: %v", err)
		return err
	}

	return nil
}
