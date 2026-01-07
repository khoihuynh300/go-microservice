package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/khoihuynh300/go-microservice/notification-service/internal/service"
	"github.com/khoihuynh300/go-microservice/shared/pkg/messaging/events"
)

const (
	VerifyEmailSubject          = "Verify Your Email"
	EmailVerifySuccessSubject   = "Email Verify Successfully"
	ResetPasswordSubject        = "Reset Your Password"
	ResetPasswordSuccessSubject = "Password Reset Successfully"
)

type UserEventHandler struct {
	emailService service.EmailService
	baseURL      string
}

func NewUserEventHandler(emailService service.EmailService, baseURL string) EventHandler {
	return &UserEventHandler{
		emailService: emailService,
		baseURL:      baseURL,
	}
}

func (h *UserEventHandler) HandleEvent(ctx context.Context, event *events.Event) error {
	switch event.EventType {
	case events.TypeUserRegisteredEvent:
		return h.handleUserRegistered(ctx, event)
	case events.TypeEmailVerifySuccessEvent:
		return h.handleEmailVerifySuccess(ctx, event)
	case events.TypeForgotPasswordEvent:
		return h.handleUserForgotPassword(ctx, event)
	case events.TypePasswordResetSuccessEvent:
		return h.handlePasswordResetSuccess(ctx, event)
	default:
		log.Printf("Unhandled event type: %s", event.EventType)
		return nil
	}
}

func (h *UserEventHandler) handleUserRegistered(ctx context.Context, event *events.Event) error {
	jsonData, err := json.Marshal(event.Data)
	if err != nil {
		return fmt.Errorf("failed to marshal event data: %w", err)
	}

	var payload events.UserRegisteredEvent
	if err := json.Unmarshal(jsonData, &payload); err != nil {
		return fmt.Errorf("invalid event data format: %w", err)
	}

	verificationLink := fmt.Sprintf("%s/verify-email?token=%s",
		h.baseURL, payload.Token)

	emailData := map[string]any{
		"Subject":          VerifyEmailSubject,
		"FullName":         payload.FullName,
		"VerificationLink": verificationLink,
	}

	if err := h.emailService.SendTemplateEmail(ctx, "verify_email", []string{payload.Email}, emailData); err != nil {
		log.Printf("Failed to send verify email: %v", err)
		return fmt.Errorf("failed to send verify email: %w", err)
	}

	log.Printf("Verify email sent successfully to: %s", payload.Email)
	return nil
}

func (h *UserEventHandler) handleEmailVerifySuccess(ctx context.Context, event *events.Event) error {
	jsonData, err := json.Marshal(event.Data)
	if err != nil {
		return fmt.Errorf("failed to marshal event data: %w", err)
	}

	var payload events.EmailVerifySuccessEvent
	if err := json.Unmarshal(jsonData, &payload); err != nil {
		return fmt.Errorf("invalid event data format: %w", err)
	}

	emailData := map[string]any{
		"Subject":      EmailVerifySuccessSubject,
		"Email":        payload.Email,
		"HomePageLink": h.baseURL,
	}

	if err := h.emailService.SendTemplateEmail(ctx, "email_verified", []string{payload.Email}, emailData); err != nil {
		log.Printf("Failed to send email verified email: %v", err)
		return fmt.Errorf("failed to send email verified email: %w", err)
	}

	log.Printf("Email verified email sent to: %s", payload.Email)
	return nil
}

func (h *UserEventHandler) handleUserForgotPassword(ctx context.Context, event *events.Event) error {
	jsonData, err := json.Marshal(event.Data)
	if err != nil {
		return fmt.Errorf("failed to marshal event data: %w", err)
	}

	var payload events.UserForgotPasswordEvent
	if err := json.Unmarshal(jsonData, &payload); err != nil {
		return fmt.Errorf("invalid event data format: %w", err)
	}

	resetLink := fmt.Sprintf("%s/reset-password?token=%s",
		h.baseURL, payload.Token)

	emailData := map[string]any{
		"Subject":   ResetPasswordSubject,
		"FullName":  payload.FullName,
		"ResetLink": resetLink,
	}

	if err := h.emailService.SendTemplateEmail(ctx, "forgot_password", []string{payload.Email}, emailData); err != nil {
		log.Printf("Failed to send password reset email: %v", err)
		return fmt.Errorf("failed to send password reset email: %w", err)
	}

	log.Printf("Password reset email sent to: %s", payload.Email)
	return nil
}

func (h *UserEventHandler) handlePasswordResetSuccess(ctx context.Context, event *events.Event) error {
	jsonData, err := json.Marshal(event.Data)
	if err != nil {
		return fmt.Errorf("failed to marshal event data: %w", err)
	}

	var payload events.UserPasswordResetSuccessEvent
	if err := json.Unmarshal(jsonData, &payload); err != nil {
		return fmt.Errorf("invalid event data format: %w", err)
	}

	emailData := map[string]any{
		"Subject":      ResetPasswordSuccessSubject,
		"Email":        payload.Email,
		"HomePageLink": h.baseURL,
	}

	if err := h.emailService.SendTemplateEmail(ctx, "password_reset_success", []string{payload.Email}, emailData); err != nil {
		log.Printf("Failed to send password reset success email: %v", err)
		return fmt.Errorf("failed to send password reset success email: %w", err)
	}

	log.Printf("Password reset success email sent to: %s", payload.Email)
	return nil
}

func (h *UserEventHandler) CanHandle(eventType string) bool {
	handledTypes := []string{
		events.TypeUserRegisteredEvent,
		events.TypeEmailVerifySuccessEvent,
		events.TypeForgotPasswordEvent,
		events.TypePasswordResetSuccessEvent,
	}

	for _, t := range handledTypes {
		if t == eventType {
			return true
		}
	}
	return false
}
