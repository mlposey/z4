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
	sig := make(chan os.Signal, 1)
	signal.Notify(sig,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	config := configFromEnv()
	if config.ProfilerEnabled {
		f, err := os.Create("./z4_cpu.profile")
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	initLogger(config.DebugLoggingEnabled)
	db := initDB(config.DBDataDir)

	srv := server.NewServer(server.Config{
		DB:            db,
		ServicePort:   config.ServicePort,
		PeerPort:      config.PeerPort,
		RaftDataDir:   config.PeerDataDir,
		BootstrapRaft: config.BootstrapCluster,
	})
	go func() {
		e := srv.Start()
		if e != nil {
			telemetry.Logger.Error("server failed", zap.Error(e))
			sig <- syscall.SIGQUIT
		}
	}()

	<-sig

	err := srv.Close()
	if err != nil {
		telemetry.Logger.Error("error stopping server", zap.Error(err))
	}

	err = db.Close()
	if err != nil {
		telemetry.Logger.Error("error closing database", zap.Error(err))
	}
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
