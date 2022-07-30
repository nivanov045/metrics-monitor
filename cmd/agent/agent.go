package main

import (
	"flag"
	"github.com/caarlos0/env/v6"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/nivanov045/silver-octo-train/cmd/agent/metricsagent"
)

func main() {
	log.Println("agent::main: started")

	var cfg metricsagent.Config
	err := env.Parse(&cfg)
	if err != nil {
		log.Fatalln("agent::main: error in env parsing:", err)
	}

	flag.StringVar(&cfg.Address, "a", cfg.Address, "address")
	log.Println(cfg.Address)
	flag.DurationVar(&cfg.PollInterval, "p", cfg.PollInterval, "poll interval")
	flag.DurationVar(&cfg.ReportInterval, "r", cfg.ReportInterval, "report interval")
	flag.Parse()

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
