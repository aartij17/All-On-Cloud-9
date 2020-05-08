package main

import (
	"All-On-Cloud-9/config"
	"All-On-Cloud-9/server/nodes"
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
		return fmt.Sprintf(config.NODE_NAME, config.APP_MANUFACTURER, nodeIdNum)
	case config.APP_SUPPLIER:
		return fmt.Sprintf(config.NODE_NAME, config.APP_SUPPLIER, nodeIdNum)
	case config.APP_CARRIER:
		return fmt.Sprintf(config.NODE_NAME, config.APP_CARRIER, nodeIdNum)
	case config.APP_BUYER:
		return fmt.Sprintf(config.NODE_NAME, config.APP_BUYER, nodeIdNum)
	}
	return ""
}

func main() {
	var (
		nodeIdNum      int
		appName        string
		configFilePath string
	)
	flag.IntVar(&nodeIdNum, "nodeId", 1, "node ID(1, 2, 3, 4)")
	flag.StringVar(&appName, "appName", config.APP_MANUFACTURER, fmt.Sprintf("apps - %s, %s, %s, %s",
		config.APP_MANUFACTURER, config.APP_BUYER, config.APP_CARRIER, config.APP_SUPPLIER))
	flag.StringVar(&configFilePath, "configFilePath", "", "")
	if configFilePath == "" {
		log.Error("invalid config file path found, exiting now")
		os.Exit(1)
	}
	ctx, cancel := context.WithCancel(context.Background())
	// load all the config details - hosts, ports, service names
	config.LoadConfig(ctx, configFilePath)
	nodeId := getNodeId(appName, nodeIdNum)
	if nodeId == "" {
		log.Error("invalid node Id created, exiting now")
		os.Exit(1)
	}
	nodes.StartServer(ctx, nodeId, appName, nodeIdNum)
	signalChan := make(chan os.Signal, 1)
	cleanupDone := make(chan bool)
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
