package telemetry

import (
	"go.uber.org/zap"
)

// Logger is the global logger.
//
// Clients should use this directly after it has
// been configured with the InitLogger method.
var Logger *zap.Logger

// InitLogger initializes Logger.
//
// Enabling debug logging will display logs in a more readable
// format in addition to displaying debug-level logs and above.
func InitLogger(debugEnabled bool) error {
	var err error
	if debugEnabled {
		Logger, err = zap.NewDevelopment()
	} else {
		Logger, err = zap.NewProduction()
	}
	return err
}

// LogRequestID returns a zap log field for http request IDs.
func LogRequestID(id string) zap.Field {
	return zap.String("request_id", id)
}
