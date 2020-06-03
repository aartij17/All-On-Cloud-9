package application

import (
	"All-On-Cloud-9/common"
	"All-On-Cloud-9/config"
	"All-On-Cloud-9/messenger"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/nats-io/nats.go"
	log "github.com/sirupsen/logrus"
)

var (
	// Define some arbitrary shipping rates for each carrier
	Rates                   = map[string]int{"USPS": 1, "FEDEX": 2, "UPS": 3}
	ManufacturerCostPerUnit = 5 // default value
	SupplierCostPerUnit     = 5 // Default value

	AppAgentChan                = make(chan *common.Message)
	sendClientRequestToAppsChan = make(chan *common.Transaction)
)

// startInterAppNatsListener is basically the server side of the applications
// handling all the incoming requests from the other applications acting like clients
func startInterAppNatsListener(ctx context.Context, msgChan chan *nats.Msg) {
	var (
		msg *common.Message
	)
	for {
		select {
		case natsMsg := <-msgChan:
			_ = json.Unmarshal(natsMsg.Data, &msg)
			//fmt.Println(msg)
			common.UpdateGlobalClock(msg.Txn.Clock.Clock, false)
			AppAgentChan <- msg
		}
	}
}

func startClient(ctx context.Context, addr string, port string, handler func(http.ResponseWriter, *http.Request)) error {
	http.HandleFunc(addr, handler)
	err := http.ListenAndServe(":"+port, nil)
	return err
}

// advertiseTransactionMessage starts a listener on messages incoming from other applications.
// so for example, supplier could act as a client and send a message to manufacturer.
// in this case, supplier will adverstise a message to NATS_MANUFACTURER_INBOX
func advertiseTransactionMessage(ctx context.Context, nc *nats.Conn,
	fromApp string, serverId string, serverNumId int) {
	var (
		txn *common.Transaction
	)
	// This requires all transaction struct to have a ToApp field
	for {
		select {
		// send the client request to the target application
		case txn = <-sendClientRequestToAppsChan:
			txn.FromId = serverId
			txn.FromApp = fromApp
			txn.ToId = fmt.Sprintf(config.NODE_NAME, txn.ToApp, 0)
			txn.FromNodeNum = serverNumId
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
				Clock:       txn.Clock,
			}
			log.WithField("currentNTime", time.Now().UnixNano()).Info("ready to send a message to nats")
			jMsg, _ := json.Marshal(msg)
			toNatsInbox := fmt.Sprintf("NATS_%s_INBOX", txn.ToApp)
			messenger.PublishNatsMessage(ctx, nc, toNatsInbox, jMsg)
		}
	}
}
