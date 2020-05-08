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
	buyer *Buyer
)

type Buyer struct {
	MsgChannel    chan *nats.Msg
	ContractValid chan bool
}

func (b *Buyer) subToInterAppNats(ctx context.Context, nc *nats.Conn) {
	var (
		err error
	)
	err = messenger.SubscribeToInbox(ctx, nc, common.NATS_BUYER_INBOX, b.MsgChannel)
	if err != nil {
		log.WithFields(log.Fields{
			"error":       err.Error(),
			"application": config.APP_BUYER,
			"topic":       common.NATS_BUYER_INBOX,
		}).Error("error subscribing to the nats topic")
	}
}

func (b *Buyer) processTxn(ctx context.Context, msg *common.Message) {

}

func StartBuyerApplication(ctx context.Context, nc *nats.Conn) {
	buyer = &Buyer{
		ContractValid: make(chan bool),
		MsgChannel:    make(chan *nats.Msg),
	}
	// all the other app-specific business logic can come here.
	buyer.subToInterAppNats(ctx, nc)
	startInterAppNatsListener(ctx, buyer.MsgChannel)
}
