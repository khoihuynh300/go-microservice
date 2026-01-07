package service

import "context"

type Email struct {
	To       []string
	Subject  string
	HTMLBody string
}

type EmailService interface {
	SendEmail(ctx context.Context, email *Email) error

	SendTemplateEmail(ctx context.Context, templateName string, to []string, data map[string]any) error
}
