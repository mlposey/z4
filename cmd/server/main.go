package main

import (
	"github.com/mlposey/z4/server"
	"github.com/mlposey/z4/server/cluster"
	"github.com/mlposey/z4/storage"
	"github.com/mlposey/z4/storage/sqli"
	"github.com/mlposey/z4/telemetry"
	"go.uber.org/zap"
	"log"
	_ "net/http/pprof"
	"os"
	"os/signal"
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

	initLogger(config.DebugLoggingEnabled)
	db := initDB(config.DBDataDir, config.SQLPort)

	srv := server.NewServer(server.Config{
		DB:          db,
		GRPCPort:    config.ServicePort,
		MetricsPort: config.MetricsPort,
		PeerConfig: cluster.PeerConfig{
			ID:               config.PeerID,
			Port:             config.PeerPort,
			AdvertiseAddr:    config.PeerAdvertiseAddr,
			DataDir:          config.RaftDataDir,
			LogBatchSize:     1000,
			BootstrapCluster: config.BootstrapCluster,
		},
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

func initDB(dataDir string, port int) *storage.PebbleClient {
	db, err := storage.NewPebbleClient(dataDir)
	if err != nil {
		log.Fatalf("error initializing database client: %v", err)
	}
	go sqli.StartWireListener(sqli.WireConfig{
		Port:       port,
		Tasks:      storage.NewTaskStore(db),
		Namespaces: storage.NewNamespaceStore(db),
	})
	return db
}
