package nodes

import (
	"All-On-Cloud-9/common"
	"All-On-Cloud-9/config"
	"All-On-Cloud-9/messenger"
	"All-On-Cloud-9/server/application"
	"All-On-Cloud-9/server/blockchain"
	"context"
	"encoding/json"
	"os"

	log "github.com/Sirupsen/logrus"

	"github.com/nats-io/nats.go"

	"github.com/hashicorp/terraform/dag"
)

var (
	AppServer         *Server
	AppServerNatsChan = make(chan *nats.Msg)
)

type Server struct {
	Id                     string                `json:"server_id"`
	AppName                string                `json:"appname"`
	ServerNumId            int                   `json:"numeric_id"`
	IsPrimaryAgent         bool                  `json:"is_primary_agent"`
	VertexMap              map[string]dag.Vertex `json:"vertex_map"`
	CurrentLocalTxnSeq     int                   `json:"current_local_txn_seq"`
	CurrentGlobalTxnSeq    int                   `json:"current_global_txn_seq"`
	NatsConn               *nats.Conn            `json:"nats_connection"`
	LocalConsensusComplete chan bool
}

func (server *Server) startNatsConsumer(ctx context.Context) {
	// this will receive the request message from other applications
	_ = messenger.SubscribeToInbox(ctx, server.NatsConn, common.NATS_ORD_REQUEST, AppServerNatsChan)

	// subscribe to the NATS inbox for messages from the BPAXOS module
	_ = messenger.SubscribeToInbox(ctx, server.NatsConn, common.NATS_CONSENSUS_DONE_MSG, AppServerNatsChan)
}

func (server *Server) initiateLocalGlobalConsensus(ctx context.Context, fromNodeId string, msg []byte) {
	// check if the request was received on a primary agent
	if server.ServerNumId != 1 {
		log.WithFields(log.Fields{
			"receiverNodeId": fromNodeId,
		}).Error("request received on a non-primary agent, no action taken")
		return
	}
	// TODO: [Aarti]: Check if the message is valid -- check the signature
	// TODO: Initiate local consensus - Make sure that true is published to LocalConsensusCompleteChannel
	<-server.LocalConsensusComplete
	// initiate global consensus
	messenger.PublishNatsMessage(ctx, server.NatsConn, common.NATS_CONSENSUS_INITIATE_MSG, msg)
}

func (server *Server) postConsensusProcessTxn(ctx context.Context, natsMsg *nats.Msg) {
	// 1. Check the contract if the transaction is legal
	// 2. Add the transaction to the blockchain
}

func (server *Server) startNatsListener(ctx context.Context) {
	var (
		natsMsg *nats.Msg
		msg     *common.Message
	)
	server.startNatsConsumer(ctx)
	for {
		select {
		case natsMsg = <-AppServerNatsChan:
			switch natsMsg.Subject {
			case common.NATS_ORD_REQUEST:
				_ = json.Unmarshal(natsMsg.Data, &msg)
				server.initiateLocalGlobalConsensus(ctx, msg.FromNodeId, natsMsg.Data)
			case common.NATS_CONSENSUS_DONE_MSG:
				server.postConsensusProcessTxn(ctx, natsMsg)
			}
		}
	}
}

func (server *Server) RunApplication(ctx context.Context, appName string) {
	switch appName {
	case config.APP_BUYER:
		application.StartBuyerApplication(ctx, server.NatsConn)
	case config.APP_CARRIER:
		application.StartCarrierApplication(ctx, server.NatsConn)
	case config.APP_SUPPLIER:
		application.StartSupplierApplication(ctx, server.NatsConn)
	case config.APP_MANUFACTURER:
		application.StartManufacturerApplication(ctx, server.NatsConn)
	}
}

func StartServer(ctx context.Context, nodeId string, appName string, id int) {
	nc, err := messenger.NatsConnect(ctx)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Error("error connecting to nats, exiting now...")
		os.Exit(1)
	}
	AppServer = &Server{
		Id:                     nodeId,
		AppName:                appName,
		ServerNumId:            id,
		IsPrimaryAgent:         false,
		VertexMap:              map[string]dag.Vertex{},
		NatsConn:               nc,
		LocalConsensusComplete: make(chan bool),
	}
	// initialize the public blockchain
	blockchain.InitBlockchain(id)
	AppServer.startNatsListener(ctx)
	// start the application for which the pod is spun up. it HAS to be a goroutine since we want this to be a
	// non-blocking call and also run some part of this code like a smart contract.
	go AppServer.RunApplication(ctx, appName)
	return
}
