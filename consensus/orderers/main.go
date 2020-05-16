package main

import (
	"All-On-Cloud-9/config"
	"All-On-Cloud-9/consensus/orderers/nodes"
	"context"
	"flag"
	"os"
	"os/signal"
	"strings"

	log "github.com/Sirupsen/logrus"
)

func configureLogger(level string) {
	log.SetOutput(os.Stderr)
	switch strings.ToLower(level) {
	case "panic":
		log.SetLevel(log.PanicLevel)
	case "debug":
		log.SetLevel(log.DebugLevel)
	case "info":
		log.SetLevel(log.InfoLevel)
	case "warning", "warn":
		log.SetLevel(log.WarnLevel)
	}
}

func main() {
	var (
		nodeId         int
		configFilePath string
		err            error
	)
	flag.IntVar(&nodeId, "nodeId", 0, "node ID(0 - leader, 1 - proposer, 2, 3, 4 - consensus)")
	flag.StringVar(&configFilePath, "configFilePath",
		"/Users/aartij17/go/src/All-On-Cloud-9/config/config.json", "")

	flag.Parse()

	if configFilePath == "" {
		log.Error("invalid config file path, exiting now")
		os.Exit(1)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cleanupDone := make(chan bool)

	config.LoadConfig(ctx, configFilePath)
	err = nodes.CreateOrderer(ctx, nodeId)
	if err != nil {
		cancel()
		cleanupDone <- true
	}

	signalChan := make(chan os.Signal, 1)

	signal.Notify(signalChan, os.Interrupt)
	go func() {
		for _ = range signalChan {
			log.Info("Received an interrupt, stopping all connections...")
			cancel()
			cleanupDone <- true
		}
	}()
	<-cleanupDone
}
