package application

import (
	"All-On-Cloud-9/common"
	"All-On-Cloud-9/config"
	"All-On-Cloud-9/messenger"
	"context"
	"fmt"
	"strconv"

	log "github.com/Sirupsen/logrus"
	"github.com/nats-io/nats.go"

	"encoding/json"
	"net/http"
)

var (
	BuyerObj *Buyer
)

const (
	BUYER_RETURN_REQUEST = "RETURN"
	BUYER_BUY_REQUEST    = "BUY"
)

type Buyer struct {
	MsgChannel    chan *nats.Msg
	ContractValid chan bool
}

type BuyerClientRequest struct {
	TxnType          string `json:"transaction_type"`
	ToApp            string `json:"to_application"`
	Type             string `json:"message_type"`
	UnitsTransferred int    `json:"units_transferred"`
	MoneyTransferred int    `json:"money_transferred"`
	ShippingService  string `json:"shipping_service"`
	ShippingCost     int    `json:"shipping_cost"`
}

func (b *Buyer) subToInterAppNats(ctx context.Context, nc *nats.Conn, serverId string, serverNumId int) {
	var (
		err error
	)
	err = messenger.SubscribeToInbox(ctx, nc, common.NATS_BUYER_INBOX, b.MsgChannel, false)
	if err != nil {
		log.WithFields(log.Fields{
			"error":       err.Error(),
			"application": config.APP_BUYER,
			"topic":       common.NATS_BUYER_INBOX,
		}).Error("error subscribing to the nats topic")
	}
}
func handleBuyerRequest(w http.ResponseWriter, r *http.Request) {
	var (
		mTxn *BuyerClientRequest
		txn  *common.Transaction
	)
	fmt.Println("HandleBuyerRequest")
	_ = json.NewDecoder(r.Body).Decode(&mTxn)
	fmt.Println(mTxn)

	jTxn, _ := json.Marshal(mTxn)

	txn = &common.Transaction{
		TxnBody: jTxn,
		FromApp: config.APP_BUYER,
		ToApp:   mTxn.ToApp,
		ToId:    "",
		FromId:  "",
		TxnType: mTxn.Type,
		Clock:   nil,
	}
	sendClientRequestToAppsChan <- txn
}

func StartBuyerApplication(ctx context.Context, nc *nats.Conn, serverId string, serverNumId int) {
	BuyerObj = &Buyer{
		ContractValid: make(chan bool),
		MsgChannel:    make(chan *nats.Msg),
	}
	go advertiseTransactionMessage(ctx, nc, config.APP_BUYER, serverId, serverNumId)
	// all the other app-specific business logic can come here.
	go startClient(ctx, "/app/buyer",
		strconv.Itoa(config.SystemConfig.AppInstance.AppBuyer.Servers[serverNumId].Port), handleBuyerRequest)
	BuyerObj.subToInterAppNats(ctx, nc, serverId, serverNumId)
	startInterAppNatsListener(ctx, BuyerObj.MsgChannel)
}
