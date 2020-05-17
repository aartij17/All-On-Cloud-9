package nodes

import (
	"All-On-Cloud-9/common"
	"All-On-Cloud-9/config"
	"All-On-Cloud-9/consensus/orderers/nodes"
	"All-On-Cloud-9/messenger"
	"All-On-Cloud-9/server/application"
	"All-On-Cloud-9/server/blockchain"
	"context"
	"encoding/json"
	"os"

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
	VertexMap              map[string]*blockchain.Vertex `json:"vertex_map"`
	CurrentLocalTxnSeq     int                           `json:"current_local_txn_seq"`
	CurrentGlobalTxnSeq    int                           `json:"current_global_txn_seq"`
	NatsConn               *nats.Conn                    `json:"nats_connection"`
	LocalConsensusComplete chan bool
	LastAddedLocalBlock    *blockchain.Vertex
	LastAddedGlobalBlock   *blockchain.Vertex
}

func (server *Server) startLocalConsensus() {
	log.Info("starting local consensus")
	// TODO: [Aarti]: This is a placeholder for now.
	server.LocalConsensusComplete <- true
}

func (server *Server) initiateLocalGlobalConsensus(ctx context.Context, fromNodeId string, msg []byte) {
	// check if the request was received on a primary agent
	if server.ServerNumId != 0 {
		log.WithFields(log.Fields{
			"receiverNodeId": fromNodeId,
		}).Error("request received on a non-primary agent, no action taken")
		return
	}
	// TODO: [Aarti]: Check if the message is valid -- check the signature
	// TODO: THIS WILL BLOCK! Initiate local consensus - Make sure that true is published to LocalConsensusCompleteChannel
	// [Aarti]: This needs to be a go routine since we want to ensure that we appropriately wait for the
	// local consensus to finish
	go server.startLocalConsensus()
	<-server.LocalConsensusComplete
	server.startGlobalConsensusProcess(ctx, msg)
}

// postConsensusProcessTxn is called once the local consensus has been reached by the nodes.
func (server *Server) startGlobalConsensusProcess(ctx context.Context, msg []byte) {
	log.Info("LET'S START THE GLOBAL CONSENSUS, HERE WE GOOOOO")
	var commonMessage *common.Message
	_ = json.Unmarshal(msg, &commonMessage)
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

	//// this will receive the request message from other applications
	//_ = messenger.SubscribeToInbox(ctx, server.NatsConn, common.NATS_ORD_REQUEST, AppServerNatsChan)

	// subscribe to the NATS inbox for messages from the BPAXOS module.
	// for orderer based consensus, this message will NEVER be sent to the server agent, instead, it will be
	// sent to the orderer primary node
	// _ = messenger.SubscribeToInbox(ctx, server.NatsConn, common.NATS_CONSENSUS_DONE_MSG, AppServerNatsChan, false)

	// subscribe to the NATS inbox to receive result of the final consensus success message
	_ = messenger.SubscribeToInbox(ctx, server.NatsConn, common.NATS_ADD_TO_BC, AppServerNatsChan, false)

	go func() {
		var (
			natsMsg *nats.Msg
			msg     *common.Message
		)
		for {
			select {
			case msg = <-application.AppAgentChan:
				log.WithFields(log.Fields{
					"fromApp": msg.FromApp,
					"toApp":   msg.ToApp,
				}).Info("recieved a message from one of the applications")
				jMsg, _ := json.Marshal(msg)
				// TODO: [Aarti]: choose which consensus has to be initiated here. Orderer based or otherwise?
				server.initiateLocalGlobalConsensus(ctx, msg.FromNodeId, jMsg)
			case natsMsg = <-AppServerNatsChan:
				switch natsMsg.Subject {
				case common.NATS_CONSENSUS_DONE_MSG:
					log.Debug("NATS_CONSENSUS_DONE_MSG_RCVD, nothing to do")
				case common.NATS_ADD_TO_BC:
					var ordererMsg *nodes.Message
					_ = json.Unmarshal(natsMsg.Data, &ordererMsg)
					//log.Info(ordererMsg.CommonMessage.Clock)
					common.UpdateGlobalClock(ordererMsg.CommonMessage.Clock.Clock, false)
					server.InitiateAddBlock(ctx, ordererMsg.CommonMessage)
				}
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

	AppServer = &Server{
		Id:                     nodeId,
		AppName:                appName,
		ServerNumId:            id,
		IsPrimaryAgent:         primaryAgent,
		VertexMap:              make(map[string]*blockchain.Vertex),
		NatsConn:               nc,
		LastAddedLocalBlock:    genesisBlock,
		LastAddedGlobalBlock:   genesisBlock,
		LocalConsensusComplete: make(chan bool),
	}

	// add the genesis block to the map
	AppServer.VertexMap[common.LAMBDA_BLOCK] = genesisBlock

	go AppServer.startNatsSubscriber(ctx)
	// start the application for which the pod is spun up. it HAS to be a goroutine since we want this to be a
	// non-blocking call and also run some part of this code like a smart contract.
	go AppServer.RunApplication(ctx, appName)
	return
}
