package agentconfig

import (
	"flag"
	"github.com/caarlos0/env/v6"
	"log"
	"time"
)

type Config struct {
	Address        string        `env:"ADDRESS"`
	ReportInterval time.Duration `env:"REPORT_INTERVAL"`
	PollInterval   time.Duration `env:"POLL_INTERVAL"`
}

func BuildConfig() (Config, error) {
	var cfg Config
	cfg.buildFromFlags()
	err := cfg.buildFromEnv()
	return cfg, err
}

func (cfg *Config) buildFromFlags() {
	flag.StringVar(&cfg.Address, "a", "127.0.0.1:8080", "address")
	flag.DurationVar(&cfg.PollInterval, "p", 1*time.Second, "poll interval")
	flag.DurationVar(&cfg.ReportInterval, "r", 10*time.Second, "report interval")
	flag.Parse()
}

func (cfg *Config) buildFromEnv() error {
	err := env.Parse(cfg)
	if err != nil {
		log.Println("agentconfig::buildFromEnv: error in env parsing:", err)
	}
	return err
}
