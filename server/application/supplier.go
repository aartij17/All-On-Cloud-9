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
	supplier *Supplier
)

type Supplier struct {
	MsgChannel    chan *nats.Msg
	ContractValid chan bool
}

func (s *Supplier) subToInterAppNats(ctx context.Context, nc *nats.Conn) {
	var (
		err error
	)
	err = messenger.SubscribeToInbox(ctx, nc, common.NATS_SUPPLIER_INBOX, s.MsgChannel)
	if err != nil {
		log.WithFields(log.Fields{
			"error":       err.Error(),
			"application": config.APP_SUPPLIER,
			"topic":       common.NATS_SUPPLIER_INBOX,
		}).Error("error subscribing to the nats topic")
	}
}

func (s *Supplier) processTxn(ctx context.Context, msg *common.Message) {

}

func StartSupplierApplication(ctx context.Context, nc *nats.Conn) {
	supplier = &Supplier{
		ContractValid: make(chan bool),
		MsgChannel:    make(chan *nats.Msg),
	}
	// all the other app-specific business logic can come here.
	supplier.subToInterAppNats(ctx, nc)
	startInterAppNatsListener(ctx, supplier.MsgChannel)
}
