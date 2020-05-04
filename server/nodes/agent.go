package nodes

import (
	"All-On-Cloud-9/common"
	"All-On-Cloud-9/config"
	"All-On-Cloud-9/messenger"
	"All-On-Cloud-9/server/blockchain"
	"context"
	"encoding/json"
	"os"

	log "github.com/Sirupsen/logrus"

	"github.com/nats-io/nats.go"

	"github.com/hashicorp/terraform/dag"
)

var (
	AppServer                 *Server
	AppServerNatsChan         = make(chan *nats.Msg)
	LocalConsensusSuccessChan = make(chan bool)
)

type Server struct {
	Id                  string                `json:"server_id"`
	AppName             string                `json:"appname"`
	ServerNumId         int                   `json:"numeric_id"`
	IsPrimaryAgent      bool                  `json:"is_primary_agent"`
	VertexMap           map[string]dag.Vertex `json:"vertex_map"`
	CurrentLocalTxnSeq  int                   `json:"current_local_txn_seq"`
	CurrentGlobalTxnSeq int                   `json:"current_global_txn_seq"`
	NatsConn            *nats.Conn            `json:"nats_connection"`
}

func (server *Server) startNatsConsumer(ctx context.Context) {
	// this will receive the request message from other applications
	err := messenger.SubscribeToInbox(ctx, server.NatsConn, common.NATS_ORD_REQUEST, AppServerNatsChan)
	if err != nil {
		log.WithFields(log.Fields{
			"error":   err.Error(),
			"subject": common.NATS_ORD_REQUEST,
		}).Error("error subscribing to nats subject")
	}
}

func (server *Server) initiateLocalConsensus(ctx context.Context, msg *common.Message) {
	// check if the request was received on a primary agent
	if server.ServerNumId != 1 {
		log.WithFields(log.Fields{
			"receiverNodeId": msg.FromNodeId,
		}).Error("request received on a non-primary agent, no action taken")
		return
	}
	// TODO: [Aarti]: Check if the message is valid -- check the signature
	// TODO: Initiate local consensus

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
				server.initiateLocalConsensus(ctx, msg)
			}
		}
	}
}

func (server *Server) RunApplication(ctx context.Context, appName string) {
	switch appName {
	case config.APP_BUYER:
		startBuyerApplication(ctx)
	case config.APP_CARRIER:
		startCarrierApplication(ctx)
	case config.APP_SUPPLIER:
		startSupplierApplication(ctx)
	case config.APP_MANUFACTURER:
		startManufacturerApplication(ctx)
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
		Id:             nodeId,
		AppName:        appName,
		ServerNumId:    id,
		IsPrimaryAgent: false,
		VertexMap:      map[string]dag.Vertex{},
		NatsConn:       nc,
	}
	// initialize the public blockchain
	blockchain.InitBlockchain(id)
	AppServer.startNatsListener(ctx)
	// start the application for which the pod is spun up. it HAS to be a goroutine since we want this to be a
	// non-blocking call and also run some part of this code like a smart contract.
	go AppServer.RunApplication(ctx, appName)
	return
}
