package nodes

import (
	"All-On-Cloud-9/common"
	"All-On-Cloud-9/config"
	"All-On-Cloud-9/messenger"
	"context"
	"encoding/json"

	log "github.com/Sirupsen/logrus"

	"github.com/nats-io/nats.go"
)

var (
	manufacturer *Manufacturer
	carrier      *Carrier
	buyer        *Buyer
	supplier     *Supplier
)

const (
	NATS_MANUFACTURER_INBOX = "NATS_MANUFACTURER_INBOX"
)

type Manufacturer struct {
	MsgChannel    chan *nats.Msg
	ContractValid chan bool
}

type Carrier struct {
	MsgChannel    chan *nats.Msg
	ContractValid chan bool
}

type Buyer struct {
	MsgChannel    chan *nats.Msg
	ContractValid chan bool
}

type Supplier struct {
	MsgChannel    chan *nats.Msg
	ContractValid chan bool
}

func startBuyerApplication(ctx context.Context) {
	buyer = &Buyer{
		ContractValid: make(chan bool),
		MsgChannel:    make(chan *nats.Msg),
	}
	// all the other app-specific business logic can come here.
}

func startCarrierApplication(ctx context.Context) {
	carrier = &Carrier{
		ContractValid: make(chan bool),
		MsgChannel:    make(chan *nats.Msg),
	}
	// all the other app-specific business logic can come here.
}

func startSupplierApplication(ctx context.Context) {
	supplier = &Supplier{
		ContractValid: make(chan bool),
		MsgChannel:    make(chan *nats.Msg),
	}
	// all the other app-specific business logic can come here.
}

func startInterAppNatsListener(ctx context.Context, msgChan chan *nats.Msg) {
	var (
		msg *common.Message
	)
	for {
		select {
		case natsMsg := <-msgChan:
			_ = json.Unmarshal(natsMsg.Data, &msg)

		}
	}
}

func startManufacturerApplication(ctx context.Context) {
	manufacturer = &Manufacturer{
		ContractValid: make(chan bool),
		MsgChannel:    make(chan *nats.Msg),
	}
	// all the other app-specific business logic can come here.
	manufacturer.subToInterAppNats(ctx)
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

func (m *Manufacturer) subToInterAppNats(ctx context.Context) {
	var (
		err error
	)
	err = messenger.SubscribeToInbox(ctx, AppServer.NatsConn, NATS_MANUFACTURER_INBOX, m.MsgChannel)
	if err != nil {
		log.WithFields(log.Fields{
			"error":       err.Error(),
			"application": config.APP_MANUFACTURER,
			"topic":       NATS_MANUFACTURER_INBOX,
		}).Error("error subscribing to the nats topic")
	}
}
