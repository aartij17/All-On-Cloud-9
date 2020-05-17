package nodes

import (
	"All-On-Cloud-9/common"
	"All-On-Cloud-9/server/blockchain"
	"context"
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/hashicorp/terraform/dag"
)

func (server *Server) InitiateAddBlock(ctx context.Context, message *common.Message) {
	var (
		newBlock *blockchain.Block
		blockId  string
		isGlobal = false
		isLocal  = false
	)

	server.MapLock.Lock()
	defer server.MapLock.Unlock()

	log.WithFields(log.Fields{
		"fromApp":  message.FromApp,
		"toApp":    message.ToApp,
		"txn type": message.Txn.TxnType,
	}).Info("message details for the new block")
	if _, OK := server.PIDMap[message.Clock.PID]; OK {
		log.WithFields(log.Fields{
			"pid": message.Clock.PID,
		}).Warn("block already present in blockchain")
		return
	}
	if message.Txn.TxnType == common.LOCAL_TXN {
		if server.LastAddedLocalBlock.V.(*blockchain.Block).Clock.Clock > message.Clock.Clock {
			log.WithFields(log.Fields{
				"lastAddedTS": server.LastAddedLocalBlock.V.(*blockchain.Block).Clock.Clock,
				"incomingTS":  message.Clock.Clock,
			}).Warn("not going to add an older local block, skipping..")
			return
		}
		if server.LastAddedLocalBlock.V.(*blockchain.Block).Clock.PID == message.Clock.PID {
			log.WithFields(log.Fields{
				"messageClockId": message.Clock.PID,
			}).Warn("local block already added to the blockchain, nothing to do")
			return
		}
	} else if message.Txn.TxnType == common.GLOBAL_TXN {
		if server.LastAddedGlobalBlock.V.(*blockchain.Block).Clock.Clock > message.Clock.Clock {
			log.WithFields(log.Fields{
				"lastAddedTS": server.LastAddedGlobalBlock.V.(*blockchain.Block).Clock.Clock,
				"incomingTS":  message.Clock.Clock,
			}).Warn("not going to add an older global block, skipping..")
			return
		}
		if server.LastAddedGlobalBlock.V.(*blockchain.Block).Clock.PID == message.Clock.PID {
			log.WithFields(log.Fields{
				"messageClockId": message.Clock.PID,
			}).Warn("global block already added to the blockchain, nothing to do")
			return
		}
	}

	// if this is a local txn, the message should be intended for the current application
	if message.Txn.TxnType == common.LOCAL_TXN && server.AppName != message.ToApp {
		log.WithFields(log.Fields{
			"current app": server.AppName,
			"block app":   message.Txn.FromApp,
		}).Info("local block received, but not intended for this application")
		return
	}
	if message.Txn.TxnType == common.LOCAL_TXN && server.AppName == message.ToApp {
		blockchain.LocalSeqNumber += 1
		blockId = fmt.Sprintf(common.LOCAL_BLOCK_NUM, server.ServerNumId,
			blockchain.LocalSeqNumber)
		isLocal = true
	} else if message.Txn.TxnType == common.GLOBAL_TXN {
		// in case of a global transaction, increment both the local and the global sequence numbers
		blockchain.LocalSeqNumber += 1
		blockchain.GlobalSeqNumber += 1
		isGlobal = true

		blockId = fmt.Sprintf(common.GLOBAL_BLOCK_NUM, server.ServerNumId, blockchain.LocalSeqNumber,
			blockchain.GlobalSeqNumber)
	}

	newBlock = &blockchain.Block{
		IsGenesis:     false,
		CryptoHash:    "",
		Transaction:   message.Txn,
		InitiatorNode: message.FromNodeId,
		Clock:         message.Clock,
	}

	if isGlobal {
		newBlock.BlockId = blockId
		newBlock.ViewType = common.GLOBAL_TXN
	} else if isLocal {
		newBlock.BlockId = blockId
		newBlock.ViewType = common.LOCAL_TXN
	}
	newVertex := &blockchain.Vertex{
		VertexId: blockId,
		V:        dag.Vertex(newBlock),
	}
	if isGlobal {
		edgeGlobal := dag.BasicEdge(dag.Vertex(newVertex), dag.Vertex(server.LastAddedGlobalBlock))
		blockchain.Blockchain.Connect(edgeGlobal)
		log.WithFields(log.Fields{
			"fromVertex": newVertex.VertexId,
			"toVertex":   server.LastAddedGlobalBlock.VertexId,
		}).Info("added new edge for global block")
	}

	blockchain.Blockchain.Add(dag.Vertex((newVertex)))
	edgeLocal := dag.BasicEdge(dag.Vertex(newVertex), dag.Vertex(server.LastAddedLocalBlock))
	blockchain.Blockchain.Connect(edgeLocal)

	server.VertexMap[blockId] = newVertex
	server.PIDMap[message.Clock.PID] = true

	if isLocal {
		server.LastAddedLocalBlock = newVertex
	} else if isGlobal {
		server.LastAddedGlobalBlock = newVertex
	}
	log.WithFields(log.Fields{
		"fromVertex": newVertex.VertexId,
		"toVertex":   server.LastAddedLocalBlock.VertexId,
		"pid":        message.Clock.PID,
	}).Info("added new edge for local block")

	blockchain.PrintBlockchain()
	return
}
