package application

import (
	"All-On-Cloud-9/common"
	"All-On-Cloud-9/config"
	"All-On-Cloud-9/messenger"
	"context"

	log "github.com/Sirupsen/logrus"

	nats "github.com/nats-io/nats.go"
)

var (
	manufacturer *Manufacturer
)

type Manufacturer struct {
	MsgChannel    chan *nats.Msg
	ContractValid chan bool
}

func (m *Manufacturer) subToInterAppNats(ctx context.Context, nc *nats.Conn) {
	var (
		err error
	)
	err = messenger.SubscribeToInbox(ctx, nc, common.NATS_MANUFACTURER_INBOX, m.MsgChannel)
	if err != nil {
		log.WithFields(log.Fields{
			"error":       err.Error(),
			"application": config.APP_MANUFACTURER,
			"topic":       common.NATS_MANUFACTURER_INBOX,
		}).Error("error subscribing to the nats topic")
	}
}

func StartManufacturerApplication(ctx context.Context, nc *nats.Conn) {
	manufacturer = &Manufacturer{
		ContractValid: make(chan bool),
		MsgChannel:    make(chan *nats.Msg),
	}
	// all the other app-specific business logic can come here.
	manufacturer.subToInterAppNats(ctx, nc)
	// following logic has to be taken care of here -
	// 1. Listen to the NATS channel
	// 2. once a message is received, send it to the main AppServer object which establishes consensus
	// 3. Once consensus is reached, a message is sent back to the manufacturer object
	// 4. Once the object receives the consensus results, and if the result is true, run the manufacturer
	//    smart contract.
	// 5. Listen to the smart contract channel as well, and if the result is false, tell the main AppServer that
	//    addition of the block to the blockchain cannot be performed.
	startInterAppNatsListener(ctx, manufacturer.MsgChannel)
}
