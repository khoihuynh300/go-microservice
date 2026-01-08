package kafka

import (
	"context"

	"go.uber.org/zap"
)

type Consumer interface {
	RegisterHandler(topic string, handler MessageHandler)
	Start(ctx context.Context, logger *zap.Logger) error
	Close() error
}
