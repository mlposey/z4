package main

import (
	"go.uber.org/zap"
	"log"
	"os"
	"strconv"
	"z4/server"
	"z4/telemetry"
)

func main() {
	initLogger()

	telemetry.Logger.Info("starting server")
	// TODO: Support configurable port.
	err := server.Start(6355, nil)
	if err != nil {
		telemetry.Logger.Error("server stopped",
			zap.Error(err))
	}
}

func initLogger() {
	// TODO: Centralize environment-based configurations.
	// We don't want os.Getenv scattered all over the place.
	debugEnabled, _ := strconv.ParseBool(os.Getenv("DEBUG_LOGGING_ENABLED"))
	err := telemetry.InitLogger(debugEnabled)
	if err != nil {
		log.Fatalf("error building logger: %v\n", err)
	}
}
