package main

import (
	"os"
	"os/signal"

	log "github.com/Sirupsen/logrus"
	"github.com/nats-io/nats.go"
)

func main() {

	natsOptions := nats.Options{
		Servers: []string{"nats://0.0.0.0:4222"},
		AllowReconnect: true,
		//MaxReconnect:   3,
		//ReconnectWait:  10,
		// TODO: these callbacks can be set in order to allow nats to perform failure handling.
		//ClosedCB:          nil,
		//DisconnectedErrCB: nil,
	}
	nc, err := natsOptions.Connect()

	if err != nil {
		log.WithFields(log.Fields{
			"error":  err.Error(),
		}).Error("error connecting to nats server")
		return
	}
	// run another routine to listen to the messages that you are expecting from the server
	go func(nc *nats.Conn) {
		for {
			_, err = nc.Subscribe("consensusReplyMessage", func(m *nats.Msg) {
				log.WithFields(log.Fields{
					"message": string(m.Data),
				}).Info("Received a reply message from server")
			})
		}
	}(nc)

	err = nc.PublishRequest("consensusMessage", "consensusReplyMessage", []byte("some request"))
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Error("error publishing the nats request")
	}


	log.Info("published request")
	signalChan := make(chan os.Signal, 1)
	cleanupDone := make(chan bool)
	signal.Notify(signalChan, os.Interrupt)
	go func() {
		for _ = range signalChan {
			log.Info("Received an interrupt, stopping all connections...")
			//cancel()
			cleanupDone <- true
		}
	}()
	<-cleanupDone

}
