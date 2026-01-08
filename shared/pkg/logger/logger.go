package zaplogger

import (
	"context"

	"github.com/khoihuynh300/go-microservice/shared/pkg/const/contextkeys"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func New(serviceName string, env string) (*zap.Logger, error) {
	var zcfg zap.Config

	if env == "DEV" {
		zcfg = zap.NewDevelopmentConfig()
		zcfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	} else {
		zcfg = zap.NewProductionConfig()
		zcfg.OutputPaths = []string{"stdout"}
		zcfg.ErrorOutputPaths = []string{"stderr"}

		enc := zap.NewProductionEncoderConfig()
		enc.EncodeTime = zapcore.ISO8601TimeEncoder
		zcfg.EncoderConfig = enc
		zcfg.InitialFields = map[string]any{
			"service.name": serviceName,
		}
	}

	return zcfg.Build(zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
}

func FromContext(ctx context.Context) *zap.Logger {
	if logger, ok := ctx.Value(contextkeys.LoggerKey).(*zap.Logger); ok {
		return logger
	}
	return zap.NewNop()
}
