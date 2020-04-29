package messenger

import (
	"All-On-Cloud-9/config"
	"All-On-Cloud-9/consensus/orderers/nodes"
	"context"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/nats-io/nats.go"
)

var timeout = time.Duration(10 * time.Second)

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

func SubscribeToInbox(ctx context.Context, nc *nats.Conn, o *nodes.Orderer, subject string) error {
	var err error
	// run another routine to listen to the messages that you are expecting from the server
	go func(nc *nats.Conn) {
		for {
			_, err = nc.Subscribe(subject, func(m *nats.Msg) {
				log.WithFields(log.Fields{
					"message":    string(m.Data),
					"receivedOn": o.Id,
				}).Info("Received a reply message from server")
				nodes.NatsOrdMessage <- m
			})
		}
	}(nc)
	return nil
}

func SendRequest(ctx context.Context, nc *nats.Conn, reqSubj string, replySubj string, message []byte) {
	err := nc.PublishRequest(reqSubj, replySubj, message)
	if err != nil {
		log.WithFields(log.Fields{
			"err":         err.Error(),
			"requestSubj": reqSubj,
		}).Error("error publishing request to NATS")
	}
	log.WithFields(log.Fields{
		"requestSubj": reqSubj,
	}).Debug("published request to NATS")
}
