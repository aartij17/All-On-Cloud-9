package application

import (
	"All-On-Cloud-9/common"
	"All-On-Cloud-9/config"
	"All-On-Cloud-9/messenger"
	"context"

	log "github.com/Sirupsen/logrus"
	"github.com/nats-io/nats.go"

	"encoding/json"
	"net/http"
)

var (
	carrier                      *Carrier
	sendCarrierRequestToAppsChan = make(chan *common.Transaction)
)

type Carrier struct {
	MsgChannel    chan *nats.Msg
	ContractValid chan bool
}

type CarrierClientRequest struct {
	ToApp string `json:"to_application"`
	Type  string `json:"request_type"`
	Fee   int    `json:"fee"`
}

func handleCarrierRequest(w http.ResponseWriter, r *http.Request) {
	var (
		mTxn *CarrierClientRequest
		txn  *common.Transaction
	)
	_ = json.NewDecoder(r.Body).Decode(&mTxn)
	jTxn, _ := json.Marshal(mTxn)

	txn = &common.Transaction{
		TxnBody: jTxn,
		FromApp: config.APP_CARRIER,
		ToApp:   mTxn.ToApp,
		ToId:    "",
		FromId:  "",
		TxnType: mTxn.Type,
		Clock:   nil,
	}
	sendCarrierRequestToAppsChan <- txn
}

func (c *Carrier) subToInterAppNats(ctx context.Context, nc *nats.Conn, serverId string, serverNumId int) {
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

	go sendTransactionMessage(ctx, nc, sendCarrierRequestToAppsChan, config.APP_CARRIER, serverId, serverNumId)

}

func (c *Carrier) processTxn(ctx context.Context, msg *common.Message) {

}

func StartCarrierApplication(ctx context.Context, nc *nats.Conn, serverId string,
	serverNumId int) {
	carrier = &Carrier{
		ContractValid: make(chan bool),
		MsgChannel:    make(chan *nats.Msg),
	}
	go startClient(ctx, "/app/carrier", "8080", handleCarrierRequest)
	// all the other app-specific business logic can come here.
	carrier.subToInterAppNats(ctx, nc, serverId, serverNumId)
	startInterAppNatsListener(ctx, carrier.MsgChannel)
}
