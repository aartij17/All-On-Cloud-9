package application

import (
	"All-On-Cloud-9/common"
	"All-On-Cloud-9/config"
	"All-On-Cloud-9/messenger"
	"All-On-Cloud-9/pbft"
	"All-On-Cloud-9/pbftSingleLayer"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	log "github.com/Sirupsen/logrus"
	"github.com/nats-io/nats.go"
)

var (
	// Define some arbitrary shipping rates for each carrier
	Rates                   = map[string]int{"USPS": 1, "FEDEX": 2, "UPS": 3}
	ManufacturerCostPerUnit = 5 // default value
	SupplierCostPerUnit     = 5 // Default value

	AppAgentChan                                          = make(chan *common.Message)
	sendClientRequestToAppsChan                           = make(chan *common.Transaction)
	nodeAppName                                           = ""
	nodeServerId                                          = -1
	pbftNode                    *pbft.PbftNode            = nil
	pbftSLNode                  *pbftSingleLayer.PbftNode = nil
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
			fmt.Println(msg.Clock)
			fmt.Println(msg)
			common.UpdateGlobalClock(msg.Clock.Clock, false)
			if msg.Clock.Clock%config.GetAppNodeCnt(nodeAppName) == nodeServerId {
				if msg.Txn.TxnType == common.GLOBAL_TXN {
					switch config.GetGlobalConsensusMethod() {
					case 1:
						AppAgentChan <- msg
					case 2:
						pbftNode.MessageIn <- *msg.Txn
					case 3:
						pbftSLNode.MessageIn <- *msg.Txn
					}
				}
			}
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
	log.Info("adverstising....")
	var (
		txn *common.Transaction
	)

	totalNodesGlobal := config.GetAppCnt()
	totalNodes := config.GetAppNodeCnt(fromApp)
	nodeAppName = fromApp
	nodeServerId = serverNumId
	pbftNode = pbft.NewPbftNode(ctx, nc, fromApp, totalNodes/3, totalNodes, totalNodesGlobal/3, totalNodesGlobal, serverNumId, config.GetAppId(fromApp))
	pbftSLNode = pbftSingleLayer.NewPbftNode(ctx, nc, fromApp, serverNumId, config.GetAppId(fromApp))
	if config.IsByzantineTolerant(nodeAppName) {
		pbft.PipeInLocalConsensus(pbftNode)
	} else {
		// should forward the msg to BPaxos and retrieve the pipe-in the response
	}
	// This requires all transaction struct to have a ToApp field
	for {
		select {
		// send the client request to the target application
		case txn = <-sendClientRequestToAppsChan:
			txn.FromId = serverId
			txn.FromApp = fromApp
			txn.ToId = fmt.Sprintf(config.NODE_NAME, txn.ToApp, 0)
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
				Clock:       txn.Clock,
			}

			jMsg, _ := json.Marshal(msg)
			toNatsInbox := fmt.Sprintf("NATS_%s_INBOX", txn.ToApp)
			log.Info("ready to send a message to nats")
			messenger.PublishNatsMessage(ctx, nc, toNatsInbox, jMsg)
		}
	}
}
