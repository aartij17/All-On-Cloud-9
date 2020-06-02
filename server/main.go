package main

import (
	"All-On-Cloud-9/common"
	"All-On-Cloud-9/config"
	"All-On-Cloud-9/server/nodes"
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"

	log "github.com/sirupsen/logrus"
)

//func configureLogger(level string) {
//	log.SetFormatter(&log.TextFormatter{
//		ForceColors:               true,
//		DisableColors:             false,
//		ForceQuote:                false,
//		DisableQuote:              false,
//		EnvironmentOverrideColors: false,
//		DisableTimestamp:          false,
//		FullTimestamp:             false,
//		TimestampFormat:           "",
//		DisableSorting:            false,
//		SortingFunc:               nil,
//		DisableLevelTruncation:    false,
//		PadLevelText:              false,
//		QuoteEmptyFields:          false,
//		FieldMap:                  nil,
//		CallerPrettyfier:          nil,
//	})
//	log.SetOutput(os.Stdout)
//	switch strings.ToLower(level) {
//	case "panic":
//		log.SetLevel(log.PanicLevel)
//	case "debug":
//		log.SetLevel(log.DebugLevel)
//	case "info":
//		log.SetLevel(log.InfoLevel)
//	case "warning", "warn":
//		log.SetLevel(log.WarnLevel)
//	}
//}

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
	flag.IntVar(&nodeIdNum, "nodeId", 0, "node ID(0, 1, 2, 3)")
	flag.StringVar(&appName, "appName", config.APP_MANUFACTURER, fmt.Sprintf("apps - %s, %s, %s, %s",
		config.APP_MANUFACTURER, config.APP_BUYER, config.APP_CARRIER, config.APP_SUPPLIER))
	flag.StringVar(&configFilePath, "configFilePath",
		"/Users/aartij17/go/src/All-On-Cloud-9/config/config.json", "")

	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	// load all the config details - hosts, ports, service names
	config.LoadConfig(ctx, configFilePath)
	common.ConfigureLogger(config.SystemConfig.LogLevel)

	if configFilePath == "" {
		log.Error("invalid config file path found, exiting now")
		os.Exit(1)
	}

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
