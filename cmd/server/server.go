package main

import (
	"log"

	"github.com/nivanov045/silver-octo-train/cmd/server/api"
	"github.com/nivanov045/silver-octo-train/cmd/server/config"
	"github.com/nivanov045/silver-octo-train/cmd/server/service"
	"github.com/nivanov045/silver-octo-train/cmd/server/storage"
)

func main() {
	cfg, err := config.BuildConfig()
	if err != nil {
		log.Fatalln("server::main::error: in env parsing:", err)
	}
	log.Println("server::main::info: cfg:", cfg)

	myStorage, err := storage.New(cfg)
	if err != nil {
		if err.Error() != `can't create database'` {
			log.Fatalln(`server::main::error: unknown error in storage creation:'`, err)
		}
		log.Println(`server::main::warning: can't create database; fallback to inmemory storage`)
		myStorage = storage.NewForcedInMemory(cfg)
	}
	serv := service.New(cfg.Key, myStorage)
	myapi := api.New(serv)
	log.Fatalln(myapi.Run(cfg.Address))
}
