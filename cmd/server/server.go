package main

import (
	"github.com/nivanov045/silver-octo-train/cmd/server/api"
	"github.com/nivanov045/silver-octo-train/cmd/server/serverconfig"
	"github.com/nivanov045/silver-octo-train/cmd/server/service"
	"github.com/nivanov045/silver-octo-train/cmd/server/storage"
	"log"
)

func main() {
	cfg, err := serverconfig.BuildConfig()
	if err != nil {
		log.Fatalln("server::main: error in env parsing:", err)
	}
	log.Println("server::main: cfg:", cfg)

	myStorage := storage.New(cfg.StoreInterval, cfg.StoreFile, cfg.Restore)
	serv := service.New(myStorage)
	myapi := api.New(serv)
	log.Fatalln(myapi.Run(cfg.Address))
}
