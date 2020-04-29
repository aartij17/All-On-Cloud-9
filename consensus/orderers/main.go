package orderers

import (
	"All-On-Cloud-9/config"
	"context"
	"os"
	"os/signal"

	log "github.com/Sirupsen/logrus"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	id := config.NODE11 // TODO: Send configurable ID here
	_ = id // TODO: remove this, added to keep compiler happy
	config.LoadConfig(ctx, "/Users/aartij17/go/src/All-On-Cloud-9/config/config.json")

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
