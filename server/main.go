package main

import (
	"All-On-Cloud-9/config"
	"All-On-Cloud-9/server/nodes"
	"context"
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
	ctx, cancel := context.WithCancel(context.Background())

	// TODO: Send configurable ID here
	id := config.NODE11
	rootNode := nodes.StartServer(ctx, id)
	_ = rootNode // TODO: Added to get rid of compilation errors, remove this
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
