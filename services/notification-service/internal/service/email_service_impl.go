package service

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/smtp"
	"strings"

	"github.com/khoihuynh300/go-microservice/notification-service/internal/template"
)

type emailService struct {
	templateParser *template.Parser
	host           string
	port           int
	username       string
	password       string
	useTLS         bool
}

func NewEmailService(
	templateParser *template.Parser,
	host string,
	port int,
	username string,
	password string,
	useTLS bool,
) EmailService {
	return &emailService{
		templateParser: templateParser,
		host:           host,
		port:           port,
		username:       username,
		password:       password,
		useTLS:         useTLS,
	}
}

func (s *emailService) SendEmail(ctx context.Context, email *Email) error {
	message := s.buildMessage(email)

	if err := s.sendSMTP(email.To, message); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

func (s *emailService) SendTemplateEmail(ctx context.Context, templateName string, to []string, data map[string]any) error {
	subject, body, err := s.templateParser.Render(templateName, data)
	if err != nil {
		return fmt.Errorf("failed to render template: %w", err)
	}

	email := &Email{
		To:       to,
		Subject:  subject,
		HTMLBody: body,
	}

	return s.SendEmail(ctx, email)
}

func (s *emailService) buildMessage(email *Email) string {
	var msg strings.Builder

	msg.WriteString(fmt.Sprintf("From: %s\r\n", s.username))
	msg.WriteString(fmt.Sprintf("To: %s\r\n", strings.Join(email.To, ",")))
	msg.WriteString(fmt.Sprintf("Subject: %s\r\n", email.Subject))
	msg.WriteString("MIME-Version: 1.0\r\n")
	msg.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
	msg.WriteString("\r\n")
	msg.WriteString(email.HTMLBody)

	return msg.String()
}

func (s *emailService) sendSMTP(to []string, message string) error {
	addr := fmt.Sprintf("%s:%d", s.host, s.port)

	var auth smtp.Auth
	if s.username != "" && s.password != "" {
		auth = smtp.PlainAuth("", s.username, s.password, s.host)
	}

	if s.useTLS {
		return s.sendWithTLS(addr, auth, to, message)
	}

	return smtp.SendMail(addr, auth, s.username, to, []byte(message))
}

func (s *emailService) sendWithTLS(addr string, auth smtp.Auth, to []string, message string) error {
	client, err := smtp.Dial(addr)
	if err != nil {
		return fmt.Errorf("failed to connect to SMTP server: %w", err)
	}
	defer client.Close()

	tlsConfig := &tls.Config{
		ServerName:         s.host,
		InsecureSkipVerify: false,
	}

	if err := client.StartTLS(tlsConfig); err != nil {
		return fmt.Errorf("failed to start TLS: %w", err)
	}

	if auth != nil {
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("authentication failed: %w", err)
		}
	}

	if err := client.Mail(s.username); err != nil {
		return fmt.Errorf("failed to set sender: %w", err)
	}

	for _, recipient := range to {
		if err := client.Rcpt(recipient); err != nil {
			return fmt.Errorf("failed to add recipient %s: %w", recipient, err)
		}
	}

	writer, err := client.Data()
	if err != nil {
		return fmt.Errorf("failed to create data writer: %w", err)
	}

	if _, err := writer.Write([]byte(message)); err != nil {
		return fmt.Errorf("failed to write message: %w", err)
	}

	if err := writer.Close(); err != nil {
		return fmt.Errorf("failed to close writer: %w", err)
	}

	return client.Quit()
}
