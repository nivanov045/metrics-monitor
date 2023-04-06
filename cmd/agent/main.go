package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/rs/zerolog/pkgerrors"

	"github.com/nivanov045/silver-octo-train/internal/agent/config"
	"github.com/nivanov045/silver-octo-train/internal/agent/metricsagent"
)

func main() {
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
	zerolog.SetGlobalLevel(zerolog.DebugLevel)

	log.Info().Msg("agent started")

	cfg, err := config.BuildConfig()
	if err != nil {
		log.Panic().Err(err).Stack()
	}
	log.Debug().Interface("config", cfg).Msg("agent config")

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc,
		syscall.SIGTERM,
		syscall.SIGINT,
		syscall.SIGQUIT)

	agent := metricsagent.New(cfg)
	agent.Start()
	<-sigc
}
