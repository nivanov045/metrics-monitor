package main

import (
	"github.com/nivanov045/silver-octo-train/cmd/agent/agentconfig"
	"github.com/nivanov045/silver-octo-train/cmd/agent/metricsagent"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	log.Println("agent::main: started")
	cfg, err := agentconfig.BuildConfig()
	if err != nil {
		log.Fatalln("agent::main: error in config build:", err)
	}
	log.Println("agent::main: cfg:", cfg)

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc,
		syscall.SIGTERM,
		syscall.SIGINT,
		syscall.SIGQUIT)
	
	agent := metricsagent.New(cfg)
	agent.Start()
	<-sigc
}
