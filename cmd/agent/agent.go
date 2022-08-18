package main

import (
	"github.com/nivanov045/silver-octo-train/cmd/agent/config"
	"github.com/nivanov045/silver-octo-train/cmd/agent/metricsagent"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	log.Println("agent::main::info: started")
	cfg, err := config.BuildConfig()
	if err != nil {
		log.Fatalln("agent::main::error:", err)
	}
	log.Println("agent::main::info: cfg:", cfg)

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc,
		syscall.SIGTERM,
		syscall.SIGINT,
		syscall.SIGQUIT)

	agent := metricsagent.New(cfg)
	agent.Start()
	<-sigc
}
