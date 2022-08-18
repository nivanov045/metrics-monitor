package main

import (
	"github.com/nivanov045/silver-octo-train/cmd/server/api"
	"github.com/nivanov045/silver-octo-train/cmd/server/config"
	"github.com/nivanov045/silver-octo-train/cmd/server/service"
	"github.com/nivanov045/silver-octo-train/cmd/server/storage"
	"log"
)

func main() {
	cfg, err := config.BuildConfig()
	if err != nil {
		log.Fatalln("server::main::error: in env parsing:", err)
	}
	log.Println("server::main::info: cfg:", cfg)

	myStorage := storage.New(cfg.StoreInterval, cfg.StoreFile, cfg.Restore, cfg.Database)
	serv := service.New(cfg.Key, myStorage)
	myapi := api.New(serv)
	log.Fatalln(myapi.Run(cfg.Address))
}
