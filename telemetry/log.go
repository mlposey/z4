package telemetry

import (
	"go.uber.org/zap"
)

var Logger *zap.Logger

func InitLogger(debugEnabled bool) error {
	var err error
	if debugEnabled {
		Logger, err = zap.NewDevelopment()
	} else {
		Logger, err = zap.NewProduction()
	}
	return err
}

func LogRequestID(id string) zap.Field {
	return zap.String("request_id", id)
}
