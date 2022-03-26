package main

import (
	"github.com/mlposey/z4/server"
	"github.com/mlposey/z4/storage"
	"github.com/mlposey/z4/telemetry"
	"go.uber.org/zap"
	"log"
)

func main() {
	config := configFromEnv()
	initLogger(config.DebugLoggingEnabled)
	db := initDB(config.DBDataDir)
	startServer(config.Port, db)
}

func initLogger(debugEnabled bool) {
	err := telemetry.InitLogger(debugEnabled)
	if err != nil {
		log.Fatalf("error building logger: %v\n", err)
	}
}

func initDB(dataDir string) *storage.BadgerClient {
	db, err := storage.NewBadgerClient(dataDir)
	if err != nil {
		log.Fatalf("error initializing database client: %v", err)
	}
	return db
}

func startServer(port int, db *storage.BadgerClient) {
	telemetry.Logger.Info("starting server")

	err := server.Start(server.Config{
		DB:   db,
		Port: port,
	})
	if err != nil {
		telemetry.Logger.Error("server stopped",
			zap.Error(err))
	}
}
