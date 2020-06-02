package nodes

import (
	"All-On-Cloud-9/common"
	"All-On-Cloud-9/config"
	"All-On-Cloud-9/messenger"
	"All-On-Cloud-9/server/application"
	"All-On-Cloud-9/server/blockchain"
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform/dag"
	log "github.com/sirupsen/logrus"
)

func (server *Server) listenToContractChannel(c chan bool, resp chan bool) {
	for {
		select {
		case r := <-c:
			resp <- r
			return
		}
	}
}

func (server *Server) InvokeSmartContract(block *blockchain.Block) bool {
	respChan := make(chan bool)
	contractValid := make(chan bool)
	log.Info("invoking smart contract")
	switch block.Transaction.FromApp {
	case config.APP_MANUFACTURER:
		go server.listenToContractChannel(contractValid, respChan)
		application.RunManufacturerContract(block, contractValid)
	case config.APP_SUPPLIER:
		go server.listenToContractChannel(contractValid, respChan)
		application.RunSupplierContract(block, contractValid)
	case config.APP_BUYER:
		go server.listenToContractChannel(contractValid, respChan)
		application.RunBuyerContract(block, contractValid)
	case config.APP_CARRIER:
		go server.listenToContractChannel(contractValid, respChan)
		application.RunCarrierContract(block, contractValid)
	}

	for {
		select {
		case r := <-respChan:
			return r
		}
	}
}

func (server *Server) InitiateAddBlock(ctx context.Context, txn *common.Transaction) {
	var (
		newBlock *blockchain.Block
		blockId  string
		isGlobal = false
		isLocal  = false
	)
	defer blockchain.PrintBlockchain()

	server.MapLock.Lock()
	defer server.MapLock.Unlock()

	log.WithFields(log.Fields{
		"fromApp":   txn.FromApp,
		"toApp":     txn.ToApp,
		"txn type":  txn.TxnType,
		"pid":       txn.Clock.PID,
		"timestamp": txn.Timestamp,
	}).Debug("message details for the new block")

	if _, OK := server.PIDMap[txn.Clock.PID]; OK {
		log.WithFields(log.Fields{
			"pid": txn.Clock.PID,
		}).Warn("block already present in blockchain")
		return
	}
	if txn.TxnType == common.LOCAL_TXN {
		if server.LastAddedLocalBlock.V.(*blockchain.Block).Clock.Clock > txn.Clock.Clock {
			log.WithFields(log.Fields{
				"lastAddedTS": server.LastAddedLocalBlock.V.(*blockchain.Block).Clock.Clock,
				"incomingTS":  txn.Clock.Clock,
			}).Warn("not going to add an older local block, skipping..")
			return
		}
		if server.LastAddedLocalBlock.V.(*blockchain.Block).Clock.PID == txn.Clock.PID {
			log.WithFields(log.Fields{
				"messageClockId": txn.Clock.PID,
			}).Warn("local block already added to the blockchain, nothing to do")
			return
		}
	} else if txn.TxnType == common.GLOBAL_TXN {
		if server.LastAddedGlobalBlock != nil &&
			server.LastAddedGlobalBlock.V.(*blockchain.Block).Clock.Clock > txn.Clock.Clock {
			log.WithFields(log.Fields{
				"lastAddedTS": server.LastAddedGlobalBlock.V.(*blockchain.Block).Clock.Clock,
				"incomingTS":  txn.Clock.Clock,
			}).Warn("not going to add an older global block, skipping..")
			return
		}
		//if server.LastAddedGlobalBlock != nil {
		//	log.WithFields(log.Fields{
		//		"pid": server.LastAddedGlobalBlock.V.(*blockchain.Block).Transaction.Clock.PID,
		//	}).Error("lets see what the previously added global block is")
		//	if server.LastAddedGlobalBlock.V.(*blockchain.Block).Transaction.Clock.PID == txn.Clock.PID {
		//		log.WithFields(log.Fields{
		//			"messageClockId": txn.Clock.PID,
		//		}).Warn("global block already added to the blockchain, nothing to do")
		//		return
		//	}
		//}
	}

	// if this is a local txn, the message should be intended for the current application
	if txn.TxnType == common.LOCAL_TXN && server.AppName != txn.ToApp {
		log.WithFields(log.Fields{
			"current app": server.AppName,
			"block app":   txn.FromApp,
		}).Info("local block received, but not intended for this application")
		return
	}
	if txn.TxnType == common.GLOBAL_TXN && txn.ToApp != server.AppName {
		log.WithFields(log.Fields{
			"currentApp": server.AppName,
			"targetApp":  txn.ToApp,
		}).Debug("global block received from consensus, target app not matching")
		return
	}

	if txn.TxnType == common.LOCAL_TXN && server.AppName == txn.ToApp {
		blockchain.LocalSeqNumber += 1
		blockId = fmt.Sprintf(common.LOCAL_BLOCK_NUM, txn.ToApp, config.GetAppId(txn.ToApp)+1,
			blockchain.LocalSeqNumber)
		isLocal = true
	} else if txn.TxnType == common.GLOBAL_TXN {
		// in case of a global transaction, increment the local sequence number ONLY
		// if the target app is the same as the currently running application
		if server.AppName == txn.ToApp {
			blockchain.LocalSeqNumber += 1
		}
		blockchain.GlobalSeqNumber += 1
		isGlobal = true

		blockId = fmt.Sprintf(common.GLOBAL_BLOCK_NUM, txn.ToApp, config.GetAppId(txn.ToApp)+1, blockchain.LocalSeqNumber,
			blockchain.GlobalSeqNumber)
	}

	newBlock = &blockchain.Block{
		IsGenesis:     false,
		CryptoHash:    "",
		Transaction:   txn,
		InitiatorNode: txn.FromId,
		Clock:         txn.Clock,
		ViewType:      txn.TxnType,
	}

	result := server.InvokeSmartContract(newBlock)
	if !result {
		log.Error("smart contract violated, skipping the block addition")
		return
	} else {
		log.Info("Contract is valid! Let's add the block already, shall we?")
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

	blockchain.Blockchain.Add(dag.Vertex((newVertex)))
	if isGlobal && server.LastAddedGlobalBlock != nil {
		edgeGlobal := dag.BasicEdge(dag.Vertex(newVertex), dag.Vertex(server.LastAddedGlobalBlock))
		blockchain.Blockchain.Connect(edgeGlobal)
		log.WithFields(log.Fields{
			"fromVertex": newVertex.VertexId,
			"toVertex":   server.LastAddedGlobalBlock.VertexId,
		}).Debug("added new edge for global block")
	}
	// if this is the first global block, add an edge b/w this block and the lambda block as well
	if isGlobal && blockchain.GlobalSeqNumber == 1 {
		lamdbaBlock := server.VertexMap[common.LAMBDA_BLOCK]
		edgeGlobal := dag.BasicEdge(dag.Vertex(newVertex), dag.Vertex(lamdbaBlock))
		blockchain.Blockchain.Connect(edgeGlobal)
	}
	// check if the new local block to be added has a previously added global block of the same app
	if txn.FromNodeNum == server.LastAddedGlobalNodeId {
		edge := dag.BasicEdge(dag.Vertex(newVertex), dag.Vertex(server.LastAddedGlobalBlock))
		blockchain.Blockchain.Connect(edge)
	}
	// irrespective of it being a local/global block, the new vertex HAS to point to
	// the last added local block. But dont connect a global block to the lcoal block if the target app
	// does not match the current app
	if !(isGlobal && server.AppName != txn.ToApp) {
		edgeLocal := dag.BasicEdge(dag.Vertex(newVertex), dag.Vertex(server.LastAddedLocalBlock))
		blockchain.Blockchain.Connect(edgeLocal)
	}

	server.VertexMap[blockId] = newVertex
	server.PIDMap[txn.Clock.PID] = true

	if isLocal {
		server.LastAddedLocalBlock = newVertex
		server.LastAddedLocalNodeId = txn.FromNodeNum
		log.WithFields(log.Fields{
			"fromVertex": newVertex.VertexId,
			"toVertex":   server.LastAddedLocalBlock.VertexId,
			"pid":        txn.Clock.PID,
		}).Debug("added new edge for local block")
	} else if isGlobal {
		server.LastAddedGlobalBlock = newVertex
		server.LastAddedGlobalNodeId = txn.FromNodeNum
		log.WithFields(log.Fields{
			"fromVertex": newVertex.VertexId,
			"toVertex":   server.LastAddedGlobalBlock.VertexId,
			"pid":        txn.Clock.PID,
		}).Debug("added new edge for global block")
		jVertex, _ := json.Marshal(newBlock)
		messenger.PublishNatsMessage(ctx, server.NatsConn, common.NATS_GLOBAL_BLOCK_INBOX, jVertex)
	}
	return
}

func (server *Server) AddForeignGlobalBlock(ctx context.Context, newBlock *blockchain.Block) {
	server.MapLock.Lock()
	defer server.MapLock.Unlock()

	if server.AppName == newBlock.Transaction.ToApp {
		log.Debug("AddForeignGlobalBlock: nothing to do")
		return
	}

	if _, OK := server.PIDMap[newBlock.Transaction.Clock.PID]; OK {
		log.WithFields(log.Fields{
			"pid": newBlock.Transaction.Clock.PID,
		}).Debug("foreign block already present in blockchain")
		return
	}

	log.WithFields(log.Fields{
		"fromApp":   newBlock.Transaction.FromApp,
		"toApp":     newBlock.Transaction.ToApp,
		"txn type":  newBlock.Transaction.TxnType,
		"pid":       newBlock.Transaction.Clock.PID,
		"timestamp": newBlock.Transaction.Timestamp,
	}).Debug("message details for the new foreign block")

	newVertex := &blockchain.Vertex{
		V:        dag.Vertex(newBlock),
		VertexId: newBlock.BlockId,
	}
	blockchain.Blockchain.Add(dag.Vertex((newVertex)))
	if server.LastAddedGlobalBlock != nil {
		edgeGlobal := dag.BasicEdge(dag.Vertex(newVertex), dag.Vertex(server.LastAddedGlobalBlock))
		blockchain.Blockchain.Connect(edgeGlobal)
		log.WithFields(log.Fields{
			"fromVertex": newVertex.VertexId,
			"toVertex":   server.LastAddedGlobalBlock.VertexId,
		}).Debug("added new edge for global block")
	}
	// if this is the first global block, add an edge b/w this block and the lambda block as well
	if blockchain.GlobalSeqNumber == 0 {
		lamdbaBlock := server.VertexMap[common.LAMBDA_BLOCK]
		edgeGlobal := dag.BasicEdge(dag.Vertex(newVertex), dag.Vertex(lamdbaBlock))
		blockchain.Blockchain.Connect(edgeGlobal)
	}
	server.LastAddedGlobalBlock = newVertex
	server.LastAddedGlobalNodeId = newBlock.Transaction.FromNodeNum
	server.VertexMap[newBlock.BlockId] = newVertex
	server.PIDMap[newBlock.Transaction.Clock.PID] = true

	log.WithFields(log.Fields{
		"fromVertex": newVertex.VertexId,
		"toVertex":   server.LastAddedGlobalBlock.VertexId,
		"pid":        newBlock.Transaction.Clock.PID,
	}).Debug("added new edge for global block")
	blockchain.GlobalSeqNumber += 1
	//blockchain.LocalSeqNumber += 1

	blockchain.PrintBlockchain()
}
