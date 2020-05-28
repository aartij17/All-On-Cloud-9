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
	ManufacturerObj *Manufacturer
)

type Manufacturer struct {
	MsgChannel    chan *nats.Msg
	ContractValid chan bool
	IsPrimary     bool
}

// subToInterAppNats subscribes to their own nats inbox to receive messages from other applications which
// may act as clients wanting to carry out a transaction
func (m *Manufacturer) subToInterAppNats(ctx context.Context, nc *nats.Conn, serverId string, serverNumId int) {
	var (
		err error
	)
	// listener: startInterAppNatsListener
	err = messenger.SubscribeToInbox(ctx, nc, common.NATS_MANUFACTURER_INBOX, m.MsgChannel, false)
	if err != nil {
		log.WithFields(log.Fields{
			"error":       err.Error(),
			"application": config.APP_MANUFACTURER,
			"topic":       common.NATS_MANUFACTURER_INBOX,
		}).Error("error subscribing to the nats topic")
	}
}

type ManufacturerClientRequest struct {
	TxnType             string `json:"transaction_type"`
	ToApp               string `json:"to_application"`
	NumUnitsToSell      int    `json:"num_units_to_sell"`
	AmountToBeCollected int    `json:"amount_to_be_collected"`
}

// handleManufacturerRequest expects a json message body containing a
// SINGLE transaction
func handleManufacturerRequest(w http.ResponseWriter, r *http.Request) {
	var (
		mTxn *ManufacturerClientRequest
		txn  *common.Transaction
	)
	//blockchain.PrintBlockchain()
	common.UpdateGlobalClock(0, false)
	id := guuid.New()
	clock := &common.LamportClock{
		PID:   fmt.Sprintf("%s_%d-%s", config.APP_MANUFACTURER, 0, id.String()),
		Clock: common.GlobalClock,
	}
	_ = json.NewDecoder(r.Body).Decode(&mTxn)
	jTxn, _ := json.Marshal(mTxn)

	log.WithFields(log.Fields{
		"request": mTxn,
		"pid":     clock.PID,
	}).Info("handling manufacturer request")

	txn = &common.Transaction{
		TxnBody:   jTxn,
		FromApp:   config.APP_MANUFACTURER,
		TxnType:   mTxn.TxnType,
		Clock:     clock,
		Timestamp: time.Now().Unix(),
	}
	if mTxn.TxnType == common.GLOBAL_TXN {
		txn.ToApp = mTxn.ToApp
	} else {
		txn.ToApp = config.APP_MANUFACTURER
	}
	sendClientRequestToAppsChan <- txn
}

func StartManufacturerApplication(ctx context.Context, nc *nats.Conn, serverId string,
	serverNumId int, isPrimary bool) {
	ManufacturerObj = &Manufacturer{
		ContractValid: make(chan bool),
		MsgChannel:    make(chan *nats.Msg),
		IsPrimary:     isPrimary,
	}
	log.WithFields(log.Fields{
		"app": config.APP_MANUFACTURER,
	}).Info("Starting the application")

	go advertiseTransactionMessage(ctx, nc, config.APP_MANUFACTURER, serverId, serverNumId)

	// has to be a go-routine cause the http handler is a blocking call
	go startClient(ctx, "/app/manufacturer",
		strconv.Itoa(config.SystemConfig.AppInstance.AppManufacturer.Servers[serverNumId].Port), handleManufacturerRequest)

	// all the other app-specific business logic can come here.
	ManufacturerObj.subToInterAppNats(ctx, nc, serverId, serverNumId)
	// following logic has to be taken care of here -
	// 1. Listen to the NATS channel
	// 2. once a message is received, send it to the main AppServer object which establishes consensus
	// 3. Once consensus is reached, a message is sent back to the manufacturer object
	// 4. Once the object receives the consensus results, and if the result is true, run the manufacturer
	//    smart contract.
	// 5. Listen to the smart contract channel as well, and if the result is false, tell the main AppServer that
	//    addition of the block to the blockchain cannot be performed.
	go startInterAppNatsListener(ctx, ManufacturerObj.MsgChannel)
}
