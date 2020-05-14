package application

import (
	"All-On-Cloud-9/common"
	"All-On-Cloud-9/config"
	"All-On-Cloud-9/messenger"
	"context"
	"encoding/json"
	"net/http"
	"fmt"
	"github.com/nats-io/nats.go"
)

var (
		// Define some arbitrary shipping rates for each carrier
	rates = map[string]int{"USPS": 1, "FEDEX": 2, "UPS": 3}
	ManufacturerCostPerUnit = 5 // default value
	SupplierCostPerUnit = 5 // Default value
)

func startInterAppNatsListener(ctx context.Context, msgChan chan *nats.Msg) {
	var (
		msg *common.Message
	)
	for {
		select {
		case natsMsg := <-msgChan:
			_ = json.Unmarshal(natsMsg.Data, &msg)
			switch msg.ToApp {
			case config.APP_MANUFACTURER:
				manufacturer.processTxn(ctx, msg)
			case config.APP_SUPPLIER:
				supplier.processTxn(ctx, msg)
			case config.APP_CARRIER:
				carrier.processTxn(ctx, msg)
			case config.APP_BUYER:
				buyer.processTxn(ctx, msg)
			}
		}
	}
}

func startClient(ctx context.Context, addr string, port string, handler func(http.ResponseWriter, *http.Request)) error {
	http.HandleFunc(addr, handler)
	err := http.ListenAndServe(":"+port, nil)
	return err
}

func sendTransactionMessage(ctx context.Context, nc *nats.Conn, channel chan *common.Transaction, fromApp string, serverId string, serverNumId int) {
	// This requires all transaction struct to have a ToApp field
	for {
			select {
			// send the client request to the target application
			case txn := <-channel:
				txn.FromId = serverId
				txn.FromApp = fromApp
				txn.ToId = fmt.Sprintf(config.NODE_NAME, txn.ToApp, 1)

				// TODO: Fill these fields correctly
				msg := common.Message{
					ToApp:       txn.ToApp,
					FromApp:     fromApp,
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
}
