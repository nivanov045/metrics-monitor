package serverconfig

import (
	"flag"
	"log"
	"time"

	"github.com/caarlos0/env/v6"
)

type Config struct {
	Address       string        `env:"ADDRESS"`
	StoreInterval time.Duration `env:"STORE_INTERVAL"`
	StoreFile     string        `env:"STORE_FILE"`
	Restore       bool          `env:"RESTORE"`
	Key           string        `env:"KEY"`
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
	flag.Parse()
}

func (cfg *Config) buildFromEnv() error {
	err := env.Parse(cfg)
	if err != nil {
		log.Println("serverconfig::buildFromEnv: error in env parsing:", err)
	}
	return err
}
