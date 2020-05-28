package nodes

import (
	"All-On-Cloud-9/common"
	"All-On-Cloud-9/config"
	"All-On-Cloud-9/consensus/orderers/nodes"
	"All-On-Cloud-9/messenger"
	"All-On-Cloud-9/pbft"
	"All-On-Cloud-9/pbftSingleLayer"
	"All-On-Cloud-9/server/application"
	"All-On-Cloud-9/server/blockchain"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"sync"

	log "github.com/Sirupsen/logrus"

	"github.com/nats-io/nats.go"
)

var (
	AppServer         *Server
	AppServerNatsChan = make(chan *nats.Msg)
)

type Server struct {
	Id                     string                        `json:"server_id"`
	AppName                string                        `json:"appname"`
	ServerNumId            int                           `json:"numeric_id"`
	IsPrimaryAgent         bool                          `json:"is_primary_agent"`
	MapLock                sync.Mutex                    `json:"lock"`
	VertexMap              map[string]*blockchain.Vertex `json:"vertex_map"`
	PIDMap                 map[string]bool               `json:"pid_map"`
	NatsConn               *nats.Conn                    `json:"nats_connection"`
	pbftNode               *pbft.PbftNode
	pbftSLNode             *pbftSingleLayer.PbftNode
	LocalConsensusComplete chan bool
	LastAddedLocalBlock    *blockchain.Vertex
	LastAddedGlobalBlock   *blockchain.Vertex
	LastAddedLocalNodeId   int
	LastAddedGlobalNodeId  int
}

func (server *Server) startLocalConsensus(commonMessage *common.Message) {
	log.Info("starting local consensus")
	server.pbftNode.MessageIn <- *commonMessage.Txn
	log.Info("message sent to channel")
	go func() {
		for {
			txn := <-server.pbftNode.MessageOut
			if !reflect.DeepEqual(txn, *commonMessage.Txn) {
				server.pbftNode.MessageOut <- txn
			} else {
				break
			}
		}
		server.LocalConsensusComplete <- true
	}()
}

func (server *Server) initiateLocalGlobalConsensus(ctx context.Context, fromNodeId string, msg []byte) {
	var commonMessage *common.Message
	_ = json.Unmarshal(msg, &commonMessage)

	// check if the request was received on a primary agent
	//if server.ServerNumId != 0 {
	//	log.WithFields(log.Fields{
	//		"receiverNodeId": fromNodeId,
	//	}).Error("request received on a non-primary agent, no action taken")
	//	return
	//}
	// TODO: [Aarti]: Check if the message is valid -- check the signature
	// TODO: THIS WILL BLOCK! Initiate local consensus - Make sure that true is published to LocalConsensusCompleteChannel
	// [Aarti]: This needs to be a go routine since we want to ensure that we appropriately wait for the
	// local consensus to finish
	fmt.Println("gonna enter IF loop")
	//if commonMessage.Clock.Clock%config.GetAppNodeCnt(server.AppName) == server.ServerNumId {
	fmt.Println("entered IF loop")
	if commonMessage.Txn.TxnType == common.GLOBAL_TXN {
		switch config.SystemConfig.GlobalConsensusAlgo {
		case common.GLOBAL_CONSENSUS_ALGO_ORDERER: //Orderer based
			server.startGlobalOrdererConsensusProcess(ctx, commonMessage)
		case common.GLOBAL_CONSENSUS_ALGO_HEIRARCHICAL: //Hierarchical PBFT
			server.pbftNode.MessageIn <- *commonMessage.Txn
			//server.pbftNode.MessageOut
		case common.GLOBAL_CONSENSUS_ALGO_SLPBFT: //Single-layer PBFT
			server.pbftSLNode.MessageIn <- *commonMessage.Txn
		}
	} else if commonMessage.Txn.TxnType == common.LOCAL_TXN {
		go server.startLocalConsensus(commonMessage) //WHY?
		<-server.LocalConsensusComplete              //WHY?
		common.UpdateGlobalClock(commonMessage.Txn.Clock.Clock, false)
		server.InitiateAddBlock(ctx, commonMessage.Txn)
		return
	}
	//}

}

// postConsensusProcessTxn is called once the local consensus has been reached by the nodes.
func (server *Server) startGlobalOrdererConsensusProcess(ctx context.Context, commonMessage *common.Message) {
	log.Info("LET'S START THE ORDERER BASED GLOBAL CONSENSUS, HERE WE GOOOOO")
	// send ORDER message to the primary of the orderer node
	message := nodes.Message{
		MessageType:   common.O_ORDER,
		Timestamp:     0,
		CommonMessage: commonMessage,
		Digest:        "",
		Hash:          "",
		FromNodeId:    server.Id,
		FromNodeNum:   server.ServerNumId,
	}
	jMsg, _ := json.Marshal(message)
	messenger.PublishNatsMessage(ctx, server.NatsConn, common.NATS_ORD_ORDER, jMsg)
}

func (server *Server) startNatsSubscriber(ctx context.Context) {
	// subscribe to the NATS inbox to receive result of the final consensus success message
	_ = messenger.SubscribeToInbox(ctx, server.NatsConn, common.NATS_ADD_TO_BC, AppServerNatsChan, false)

	go func() {
		var (
			natsMsg *nats.Msg
			msg     *common.Message
			txn     common.Transaction
		)
		for {
			select {
			// this is where the messages from the application server end up and ultimately forwarded to
			// the consensus modules
			case msg = <-application.AppAgentChan:
				log.WithFields(log.Fields{
					"fromApp": msg.FromApp,
					"toApp":   msg.ToApp,
				}).Info("received a message from one of the applications")
				jMsg, _ := json.Marshal(msg)
				server.initiateLocalGlobalConsensus(ctx, msg.FromNodeId, jMsg)
			case natsMsg = <-AppServerNatsChan:
				switch natsMsg.Subject {
				case common.NATS_CONSENSUS_DONE_MSG:
					log.Info("NATS_CONSENSUS_DONE_MSG_RCVD, nothing to do")
				case common.NATS_ADD_TO_BC:
					var ordererMsg *nodes.Message
					_ = json.Unmarshal(natsMsg.Data, &ordererMsg)
					//log.Info(ordererMsg.CommonMessage.Clock)
					common.UpdateGlobalClock(ordererMsg.CommonMessage.Txn.Clock.Clock, false)
					server.InitiateAddBlock(ctx, ordererMsg.CommonMessage.Txn)
				}
			case txn = <-server.pbftNode.MessageOut:
				log.WithFields(log.Fields{
					"type": txn.TxnType,
					"PID": txn.Clock.PID,
				}).Info("received message out from pbft")
				common.UpdateGlobalClock(txn.Clock.Clock, false)
				server.InitiateAddBlock(ctx, &txn)
			case txn = <-server.pbftSLNode.MessageOut:
				log.WithFields(log.Fields{
					"type": txn.TxnType,
					"PID": txn.Clock.PID,
				}).Info("received message out from slpbft")
				common.UpdateGlobalClock(txn.Clock.Clock, false)
				server.InitiateAddBlock(ctx, &txn)
			}
		}
	}()
}

func (server *Server) RunApplication(ctx context.Context, appName string) {
	switch appName {
	case config.APP_BUYER:
		application.StartBuyerApplication(ctx, server.NatsConn, server.Id,
			server.ServerNumId)
	case config.APP_CARRIER:
		application.StartCarrierApplication(ctx, server.NatsConn, server.Id,
			server.ServerNumId)
	case config.APP_SUPPLIER:
		application.StartSupplierApplication(ctx, server.NatsConn, server.Id,
			server.ServerNumId)
	case config.APP_MANUFACTURER:
		application.StartManufacturerApplication(ctx, server.NatsConn, server.Id,
			server.ServerNumId, server.IsPrimaryAgent)
	}
}

func StartServer(ctx context.Context, nodeId string, appName string, id int) {
	var (
		primaryAgent = false
	)
	nc, err := messenger.NatsConnect(ctx)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Error("error connecting to nats, exiting now...")
		os.Exit(1)
	}
	if id == 0 {
		primaryAgent = true
	}
	// initialize the blockchain
	genesisBlock := blockchain.InitBlockchain(nodeId)
	totalNodesGlobal := config.GetAppCnt()
	totalNodes := config.GetAppNodeCnt(appName)

	AppServer = &Server{
		Id:             nodeId,
		AppName:        appName,
		ServerNumId:    id,
		IsPrimaryAgent: primaryAgent,
		VertexMap:      make(map[string]*blockchain.Vertex),
		PIDMap:         make(map[string]bool),
		NatsConn:       nc,
		pbftNode: pbft.NewPbftNode(ctx, nc, appName, totalNodes/3, totalNodes, totalNodesGlobal/3,
			totalNodesGlobal, id, config.GetAppId(appName)),
		pbftSLNode:             pbftSingleLayer.NewPbftNode(ctx, nc, appName, id, config.GetAppId(appName)),
		LastAddedLocalBlock:    genesisBlock,
		LastAddedGlobalBlock:   nil,
		LastAddedLocalNodeId:   -1,
		LastAddedGlobalNodeId:  -1,
		LocalConsensusComplete: make(chan bool),
	}
	go pbft.PipeInHierarchicalLocalConsensus(AppServer.pbftNode) // use pbft for local consensus as well

	// add the genesis block to the map
	AppServer.VertexMap[common.LAMBDA_BLOCK] = genesisBlock

	go AppServer.startNatsSubscriber(ctx)
	// start the application for which the pod is spun up. it HAS to be a goroutine since we want this to be a
	// non-blocking call and also run some part of this code like a smart contract.
	go AppServer.RunApplication(ctx, appName)
	return
}
