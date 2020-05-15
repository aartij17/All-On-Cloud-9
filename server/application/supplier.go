package application

import (
	"All-On-Cloud-9/common"
	"All-On-Cloud-9/config"
	"All-On-Cloud-9/messenger"
	"context"
	"encoding/json"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/nats-io/nats.go"
	"net/http"
	"strconv"
)

var (
	supplier                      *Supplier
	sendSupplierRequestToAppsChan = make(chan *common.Transaction)
)

type Supplier struct {
	MsgChannel    chan *nats.Msg
	ContractValid chan bool
}

type SupplierRequest struct {
	ToApp           string `json:"to_application"`
	NumUnitsToBuy   int    `json:"num_units_to_buy"`
	AmountPaid      int    `json:"amount_paid"`
	ShippingService string `json:"shipping_service"`
	ShippingCost    int    `json:"shipping_cost"`
}

func (s *Supplier) subToInterAppNats(ctx context.Context, nc *nats.Conn, serverId string, serverNumId int) {
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

	go sendTransactionMessage(ctx, nc, sendSupplierRequestToAppsChan, config.APP_SUPPLIER, serverId, serverNumId)

}

func (s *Supplier) processTxn(ctx context.Context, msg *common.Message) {

}

func handleSupplierRequest(w http.ResponseWriter, r *http.Request) {
	var (
		sTxn *SupplierRequest
		txn  *common.Transaction
		jTxn []byte
		err  error
	)
	fmt.Println("HandleSupplierRequest")
	_ = json.NewDecoder(r.Body).Decode(&sTxn)
	jTxn, err = json.Marshal(sTxn)
	fmt.Println(sTxn)

	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Error("error handleSupplierRequest")
	}

	txn = &common.Transaction{
		TxnBody: jTxn,
		FromApp: config.APP_SUPPLIER,
		ToApp:   sTxn.ToApp,
		ToId:    "",
		FromId:  "",
		TxnType: "",
		Clock:   nil,
	}
	sendSupplierRequestToAppsChan <- txn
}

func StartSupplierApplication(ctx context.Context, nc *nats.Conn, serverId string,
	serverNumId int) {
	supplier = &Supplier{
		ContractValid: make(chan bool),
		MsgChannel:    make(chan *nats.Msg),
	}
	go startClient(ctx, "/app/supplier",
		strconv.Itoa(config.SystemConfig.AppInstance.AppSupplier.Servers[serverNumId].Port), handleSupplierRequest)
	// all the other app-specific business logic can come here.
	supplier.subToInterAppNats(ctx, nc, serverId, serverNumId)
	startInterAppNatsListener(ctx, supplier.MsgChannel)
}
