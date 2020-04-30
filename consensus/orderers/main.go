package main

import (
	"All-On-Cloud-9/config"
	"All-On-Cloud-9/consensus/orderers/nodes"
	"context"
	"flag"
	"fmt"
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

func getNodeId(appName string, nodeIdNum int) string {
	switch appName {
	case config.APP_MANUFACTURER:
		return fmt.Sprintf(config.MA_NODE, nodeIdNum)
	case config.APP_SUPPLIER:
		return fmt.Sprintf(config.S_NODE, nodeIdNum)
	case config.APP_MIDDLEMAN:
		return fmt.Sprintf(config.MI_NODE, nodeIdNum)
	case config.APP_CARRIER:
		return fmt.Sprintf(config.C_NODE, nodeIdNum)
	case config.APP_BUYER:
		return fmt.Sprintf(config.B_NODE, nodeIdNum)
	}
	return ""
}

func main() {
	var (
		nodeId         int
		configFilePath string
		err            error
	)
	flag.IntVar(&nodeId, "nodeId", 1, "node ID(1, 2, 3, 4)")
	flag.StringVar(&configFilePath, "configFilePath", "", "")
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
