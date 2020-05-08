package application

import (
	"All-On-Cloud-9/common"
	"All-On-Cloud-9/config"
	"All-On-Cloud-9/messenger"
	"context"

	log "github.com/Sirupsen/logrus"
	"github.com/nats-io/nats.go"
)

var (
	carrier *Carrier
)

type Carrier struct {
	MsgChannel    chan *nats.Msg
	ContractValid chan bool
}

func (c *Carrier) subToInterAppNats(ctx context.Context, nc *nats.Conn) {
	var (
		err error
	)
	err = messenger.SubscribeToInbox(ctx, nc, common.NATS_CARRIER_INBOX, c.MsgChannel)
	if err != nil {
		log.WithFields(log.Fields{
			"error":       err.Error(),
			"application": config.APP_CARRIER,
			"topic":       common.NATS_CARRIER_INBOX,
		}).Error("error subscribing to the nats topic")
	}
}

func (c *Carrier) processTxn(ctx context.Context, msg *common.Message) {

}

func StartCarrierApplication(ctx context.Context, nc *nats.Conn) {
	carrier = &Carrier{
		ContractValid: make(chan bool),
		MsgChannel:    make(chan *nats.Msg),
	}
	// all the other app-specific business logic can come here.
	carrier.subToInterAppNats(ctx, nc)
	startInterAppNatsListener(ctx, carrier.MsgChannel)
}
