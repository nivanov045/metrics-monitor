package main

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/rs/zerolog/pkgerrors"

	"github.com/nivanov045/silver-octo-train/internal/server/api"
	"github.com/nivanov045/silver-octo-train/internal/server/config"
	"github.com/nivanov045/silver-octo-train/internal/server/service"
	"github.com/nivanov045/silver-octo-train/internal/server/storage"
)

func main() {
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	cfg, err := config.BuildConfig()
	if err != nil {
		log.Panic().Err(err).Stack()
	}
	log.Debug().Interface("cfg", cfg).Msg("server config")

	myStorage, err := storage.New(cfg)
	if err != nil {
		if err.Error() != `can't create database'` {
			log.Error().Err(err).Stack()
		}
		log.Info().Msg(`can't create database; fallback to inmemory storage`)
		myStorage = storage.NewForcedInMemory(cfg)
	}

	serv := service.New(cfg.Key, myStorage)

	myapi := api.New(serv)

	log.Panic().Err(myapi.Run(cfg.Address))
}
