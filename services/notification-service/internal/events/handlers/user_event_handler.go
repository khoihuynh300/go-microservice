package handlers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/khoihuynh300/go-microservice/notification-service/internal/service"
	zaplogger "github.com/khoihuynh300/go-microservice/shared/pkg/logger"
	"github.com/khoihuynh300/go-microservice/shared/pkg/messaging/events"
	"go.uber.org/zap"
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
	logger := zaplogger.FromContext(ctx)

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
		logger.Warn("Unhandled event type", zap.String("event_type", event.EventType))
		return nil
	}
}

func (h *UserEventHandler) handleUserRegistered(ctx context.Context, event *events.Event) error {
	logger := zaplogger.FromContext(ctx)

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
		logger.Error("Failed to send verify email", zap.Error(err))
		return fmt.Errorf("failed to send verify email: %w", err)
	}

	logger.Info("User registered event handled successfully", zap.String("email", payload.Email))
	return nil
}

func (h *UserEventHandler) handleEmailVerifySuccess(ctx context.Context, event *events.Event) error {
	logger := zaplogger.FromContext(ctx)

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
		logger.Error("Failed to send email verified email", zap.Error(err))
		return fmt.Errorf("failed to send email verified email: %w", err)
	}

	logger.Info("Email verify success event handled successfully", zap.String("email", payload.Email))
	return nil
}

func (h *UserEventHandler) handleUserForgotPassword(ctx context.Context, event *events.Event) error {
	logger := zaplogger.FromContext(ctx)

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
		logger.Error("Failed to send password reset email", zap.Error(err))
		return fmt.Errorf("failed to send password reset email: %w", err)
	}

	logger.Info("User forgot password event handled successfully", zap.String("email", payload.Email))
	return nil
}

func (h *UserEventHandler) handlePasswordResetSuccess(ctx context.Context, event *events.Event) error {
	logger := zaplogger.FromContext(ctx)

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
		logger.Error("Failed to send password reset success email", zap.Error(err))
		return fmt.Errorf("failed to send password reset success email: %w", err)
	}

	logger.Info("Password reset success event handled successfully", zap.String("email", payload.Email))
	return nil
}
