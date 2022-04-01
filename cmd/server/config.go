package main

import (
	"github.com/kelseyhightower/envconfig"
	"log"
)

type Config struct {
	DBDataDir           string `envconfig:"DB_DATA_DIR" default:"./z4data"`
	DebugLoggingEnabled bool   `envconfig:"DEBUG_LOGGING_ENABLED" default:"false"`
	Port                int    `envconfig:"PORT" default:"6355"`
	ProfilerEnabled     bool   `envconfig:"PROFILER_ENABLED" default:"false"`
}

func configFromEnv() *Config {
	var config Config
	err := envconfig.Process("Z4", &config)
	if err != nil {
		log.Fatalf("failed to load config from environment: %v", err)
	}
	return &config
}
