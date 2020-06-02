package application

import (
	"All-On-Cloud-9/common"
	"All-On-Cloud-9/config"
	"All-On-Cloud-9/messenger"
	"context"
	"fmt"
	"strconv"

	log "github.com/sirupsen/logrus"
	"github.com/nats-io/nats.go"

	"encoding/json"
	"net/http"
)

var (
	CarrierObj *Carrier
)

type Carrier struct {
	MsgChannel    chan *nats.Msg
	ContractValid chan bool
}

type CarrierClientRequest struct {
	TxnType string `json:"transaction_type"`
	ToApp   string `json:"to_application"`
	Fee     int    `json:"fee"`
}

func handleCarrierRequest(w http.ResponseWriter, r *http.Request) {
	var (
		mTxn *CarrierClientRequest
		txn  *common.Transaction
	)
	fmt.Println("HandleCarrierRequest")
	_ = json.NewDecoder(r.Body).Decode(&mTxn)
	jTxn, _ := json.Marshal(mTxn)
	fmt.Println(mTxn)

	txn = &common.Transaction{
		TxnBody: jTxn,
		FromApp: config.APP_CARRIER,
		ToApp:   mTxn.ToApp,
		ToId:    "",
		FromId:  "",
		TxnType: mTxn.TxnType,
		Clock:   nil,
	}
	sendClientRequestToAppsChan <- txn
}

func (c *Carrier) subToInterAppNats(ctx context.Context, nc *nats.Conn, serverId string, serverNumId int) {
	var (
		err error
	)
	err = messenger.SubscribeToInbox(ctx, nc, common.NATS_CARRIER_INBOX, c.MsgChannel, false)
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

func StartCarrierApplication(ctx context.Context, nc *nats.Conn, serverId string,
	serverNumId int) {
	CarrierObj = &Carrier{
		ContractValid: make(chan bool),
		MsgChannel:    make(chan *nats.Msg),
	}
	go advertiseTransactionMessage(ctx, nc, config.APP_CARRIER, serverId, serverNumId)
	go startClient(ctx, "/app/carrier",
		strconv.Itoa(config.SystemConfig.AppInstance.AppCarrier.Servers[serverNumId].Port), handleCarrierRequest)
	// all the other app-specific business logic can come here.
	CarrierObj.subToInterAppNats(ctx, nc, serverId, serverNumId)
	startInterAppNatsListener(ctx, CarrierObj.MsgChannel)
}
