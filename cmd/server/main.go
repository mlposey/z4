package main

import (
	"github.com/mlposey/z4/server"
	"github.com/mlposey/z4/storage"
	"github.com/mlposey/z4/telemetry"
	"go.uber.org/zap"
	"log"
	"os"
	"os/signal"
	"runtime/pprof"
	"syscall"
)

func main() {
	f, err := os.Create("./cpuprofile")
	if err != nil {
		log.Fatal(err)
	}
	pprof.StartCPUProfile(f)
	defer pprof.StopCPUProfile()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	go func() {
		config := configFromEnv()
		initLogger(config.DebugLoggingEnabled)
		db := initDB(config.DBDataDir)
		startServer(config.Port, db)
		sig <- syscall.SIGQUIT
	}()

	<-sig
	// TODO: Gracefully stop server.
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
