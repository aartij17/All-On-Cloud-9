package messenger

import (
	"All-On-Cloud-9/config"
	"context"

	log "github.com/Sirupsen/logrus"

	"github.com/nats-io/nats.go"
)

func NatsConnect(ctx context.Context) (*nats.Conn, error) {
	natsOptions := nats.Options{
		Servers:        config.SystemConfig.Nats.Servers,
		AllowReconnect: true,
	}
	nc, err := natsOptions.Connect()
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Error("error connecting to nats server")
		return nil, err
	}
	return nc, nil
}

func SubscribeToInbox(ctx context.Context, nc *nats.Conn, subject string, messageChannel chan *nats.Msg) error {
	var err error
	// run another routine to listen to the messages that you are expecting from the server
	go func(nc *nats.Conn) {
		for {
			_, err = nc.Subscribe(subject, func(m *nats.Msg) {
				log.WithFields(log.Fields{
					"message": string(m.Data),
				}).Info("Received a message from NATS")
				messageChannel <- m
			})
		}
	}(nc)
	return nil
}

// PublishNatsMessage sends a NATS message to the specified NATS inbox.
func PublishNatsMessage(ctx context.Context, nc *nats.Conn, reqSubj string, message []byte) {
	var (
		err error
	)
	err = nc.PublishRequest(reqSubj, "", message)
	if err != nil {
		log.WithFields(log.Fields{
			"err":            err.Error(),
			"requestSubject": reqSubj,
		}).Error("error publishing request to NATS")
	}
	log.WithFields(log.Fields{
		"requestSubj": reqSubj,
	}).Debug("published request to NATS")
}
