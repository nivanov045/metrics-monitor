package config

import (
	"flag"
	"time"

	"github.com/caarlos0/env/v6"
	"github.com/rs/zerolog/log"
)

type Config struct {
	Address       string        `env:"ADDRESS"`
	StoreInterval time.Duration `env:"STORE_INTERVAL"`
	StoreFile     string        `env:"STORE_FILE"`
	Restore       bool          `env:"RESTORE"`
	Key           string        `env:"KEY"`
	Database      string        `env:"DATABASE_DSN"`
}

func BuildConfig() (Config, error) {
	var cfg Config
	cfg.buildFromFlags()
	err := cfg.buildFromEnv()
	return cfg, err
}

func (cfg *Config) buildFromFlags() {
	flag.StringVar(&cfg.Address, "a", "127.0.0.1:8080", "address")
	flag.DurationVar(&cfg.StoreInterval, "i", 300*time.Second, "store interval")
	flag.BoolVar(&cfg.Restore, "r", true, "restore")
	flag.StringVar(&cfg.StoreFile, "f", "/tmp/devops-metrics-db.json", "store file")
	flag.StringVar(&cfg.Key, "k", "", "key")
	flag.StringVar(&cfg.Database, "d", "", "database dsn")
	flag.Parse()
}

func (cfg *Config) buildFromEnv() error {
	err := env.Parse(cfg)
	if err != nil {
		log.Error().Err(err).Stack()
	}

	return err
}
