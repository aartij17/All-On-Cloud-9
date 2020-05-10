package application

import (
	"All-On-Cloud-9/common"
	"All-On-Cloud-9/config"
	"All-On-Cloud-9/messenger"
	"context"
	"net/http"
	"encoding/json"

	log "github.com/Sirupsen/logrus"
	"github.com/nats-io/nats.go"
)

const (
	SUPPLIER_COST_PER_UNIT = 5
)

var (
	supplier *Supplier
	sendSupplierRequestToAppsChan = make(chan *common.Transaction)
)

type Supplier struct {
	MsgChannel    chan *nats.Msg
	ContractValid chan bool
}

type SupplierRequest struct {
	ToApp               string `json:"to_application"`
	RequestType string `json:"request_type"`
	BuyRequest SupplierBuyRequest `json:"buy_request,omitempty"`
	SellRequest SupplierSellRequest `json:"sell_request,omitempty"`
}

type SupplierBuyRequest struct {
	// Supplier needs to purchase units from manufacturer
	NumUnitsToBuy      int    `json:"num_units_to_buy"`
	AmountPaid  int    `json:"amount_paid"`
}

type SupplierSellRequest struct {
	// Supplier needs to sell units to buyer
	// Supplier also needs to pay the carrier
	NumUnitsToSell     int    `json:"num_units_to_buy"`
	AmountPaid  int    `json:"amount_paid"`
	ShippingService string `json:"shipping_service"`
	ShippingCost int `json:"shipping_cost"`
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

func handleSupplierRequest(w http.ResponseWriter, r *http.Request) {
	var (
		sTxn *SupplierRequest
		txn *common.Transaction
		jTxn []byte
		err error
	)

	_ = json.NewDecoder(r.Body).Decode(&sTxn)
	switch sTxn.RequestType {
	case common.BUY_REQUEST_TYPE:
		jTxn, err = json.Marshal(sTxn.BuyRequest)
	case common.SELL_REQUEST_TYPE:
		jTxn, err = json.Marshal(sTxn.SellRequest)
	}
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
		TxnType: sTxn.RequestType,
		Clock:   nil,
	}
	sendSupplierRequestToAppsChan <- txn
}

func StartSupplierApplication(ctx context.Context, nc *nats.Conn) {
	supplier = &Supplier{
		ContractValid: make(chan bool),
		MsgChannel:    make(chan *nats.Msg),
	}
	go startClient(ctx, "/app/supplier", "8080", handleSupplierRequest)
	// all the other app-specific business logic can come here.
	supplier.subToInterAppNats(ctx, nc)
	startInterAppNatsListener(ctx, supplier.MsgChannel)
}
