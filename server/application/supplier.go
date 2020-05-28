package application

import (
	"All-On-Cloud-9/common"
	"All-On-Cloud-9/config"
	"All-On-Cloud-9/messenger"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	guuid "github.com/google/uuid"

	log "github.com/Sirupsen/logrus"
	"github.com/nats-io/nats.go"
)

var (
	SupplierObj *Supplier
)

type Supplier struct {
	MsgChannel    chan *nats.Msg
	ContractValid chan bool
}

type SupplierRequest struct {
	TxnType         string `json:"transaction_type"`
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
	err = messenger.SubscribeToInbox(ctx, nc, common.NATS_SUPPLIER_INBOX, s.MsgChannel, false)
	if err != nil {
		log.WithFields(log.Fields{
			"error":       err.Error(),
			"application": config.APP_SUPPLIER,
			"topic":       common.NATS_SUPPLIER_INBOX,
		}).Error("error subscribing to the nats topic")
	}
}

func handleSupplierRequest(w http.ResponseWriter, r *http.Request) {
	var (
		sTxn *SupplierRequest
		txn  *common.Transaction
		jTxn []byte
		err  error
	)
	common.UpdateGlobalClock(0, false)
	id := guuid.New()
	clock := &common.LamportClock{
		PID:   fmt.Sprintf("%s_%d-%s", config.APP_SUPPLIER, 0, id.String()),
		Clock: common.GlobalClock,
	}
	//blockchain.PrintBlockchain()

	_ = json.NewDecoder(r.Body).Decode(&sTxn)
	jTxn, err = json.Marshal(sTxn)
	fmt.Println(sTxn)

	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Error("error handleSupplierRequest")
		return
	}
	log.WithFields(log.Fields{
		"request": sTxn,
		"pid":     clock.PID,
	}).Info("handling supplier request")

	txn = &common.Transaction{
		TxnBody:   jTxn,
		FromApp:   config.APP_SUPPLIER,
		ToApp:     sTxn.ToApp,
		ToId:      "",
		FromId:    "",
		TxnType:   sTxn.TxnType,
		Clock:     clock,
		Timestamp: time.Now().Unix(),
	}
	if sTxn.TxnType == common.GLOBAL_TXN {
		txn.ToApp = sTxn.ToApp
	} else {
		txn.ToApp = config.APP_SUPPLIER
	}
	log.Info("about to enter -- sendSupplierRequestToAppsChan <- txn")
	sendClientRequestToAppsChan <- txn
}

func StartSupplierApplication(ctx context.Context, nc *nats.Conn, serverId string,
	serverNumId int) {
	SupplierObj = &Supplier{
		ContractValid: make(chan bool),
		MsgChannel:    make(chan *nats.Msg),
	}
	go advertiseTransactionMessage(ctx, nc, config.APP_SUPPLIER, serverId, serverNumId)
	time.Sleep(4 * time.Second)
	go startClient(ctx, "/app/supplier",
		strconv.Itoa(config.SystemConfig.AppInstance.AppSupplier.Servers[serverNumId].Port), handleSupplierRequest)
	// all the other app-specific business logic can come here.
	SupplierObj.subToInterAppNats(ctx, nc, serverId, serverNumId)
	startInterAppNatsListener(ctx, SupplierObj.MsgChannel)
}
