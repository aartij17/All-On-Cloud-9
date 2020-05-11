package application

import (
	"All-On-Cloud-9/common"
	"All-On-Cloud-9/config"
	"All-On-Cloud-9/messenger"
	"context"

	log "github.com/Sirupsen/logrus"
	"github.com/nats-io/nats.go"

	"encoding/json"
	"fmt"
	"net/http"
)

var (
	carrier *Carrier

	// Define some arbitrary shipping rates for each carrier
	rates                        = map[string]int{USPS: 1, FEDEX: 2, UPS: 3}
	sendCarrierRequestToAppsChan = make(chan *common.Transaction)
)

// Define strings for carrier names
const (
	USPS  = "USPS"
	FEDEX = "FEDEX"
	UPS   = "UPS"
)

type Carrier struct {
	MsgChannel    chan *nats.Msg
	ContractValid chan bool
}

type CarrierClientRequest struct {
	ToApp           string `json:"to_application"`
	ShippingService string `json:"shipping_service"`
	ShippingCost    int    `json:"shipping_cost"`
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
		TxnType: "",
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
	go func() {
		for {
			select {
			// send the client request to the target application
			case txn := <-sendCarrierRequestToAppsChan:
				txn.FromId = serverId
				txn.FromApp = config.APP_CARRIER
				txn.ToId = fmt.Sprintf(config.NODE_NAME, txn.ToApp, 1)

				// TODO: Fill these fields correctly
				msg := common.Message{
					ToApp:       txn.ToApp,
					FromApp:     config.APP_CARRIER,
					MessageType: "",
					Timestamp:   0,
					FromNodeId:  serverId,
					FromNodeNum: serverNumId,
					Txn:         txn,
					Digest:      "",
					PKeySig:     "",
				}

				jMsg, _ := json.Marshal(msg)
				toNatsInbox := fmt.Sprintf("NATS_%s_INBOX", txn.ToApp)
				messenger.PublishNatsMessage(ctx, nc, toNatsInbox, jMsg)
			}
		}
	}()
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
