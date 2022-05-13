package main

import (
	"github.com/kelseyhightower/envconfig"
	"log"
)

type Config struct {
	DBDataDir           string `envconfig:"DB_DATA_DIR" default:"./z4data"`
	PeerDataDir         string `envconfig:"PEER_DATA_DIR" default:"./z4peer"`
	SQLPort             int    `envconfig:"SQL_PORT" default:"3306"`
	BootstrapCluster    bool   `envconfig:"BOOTSTRAP_CLUSTER" default:"false"`
	DebugLoggingEnabled bool   `envconfig:"DEBUG_LOGGING_ENABLED" default:"false"`
	ServicePort         int    `envconfig:"SERVICE_PORT" default:"6355"`
	PeerPort            int    `envconfig:"PEER_PORT" default:"6356"`
	PeerID              string `envconfig:"PEER_ID" required:"true"`
	PeerAdvertiseAddr   string `envconfig:"PEER_ADVERTISE_ADDR" default:"127.0.0.1:6356"`
	MetricsPort         int    `envconfig:"METRICS_PORT" default:"2112"`
}

func configFromEnv() *Config {
	var config Config
	err := envconfig.Process("Z4", &config)
	if err != nil {
		log.Fatalf("failed to load config from environment: %v", err)
	}
	return &config
}
