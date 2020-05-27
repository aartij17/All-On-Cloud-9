package pbftSingleLayer

import (
	"All-On-Cloud-9/common"
	"All-On-Cloud-9/config"
	"All-On-Cloud-9/messenger"
	"context"
	"encoding/json"
	"strconv"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/nats-io/nats.go"
)

type PbftNode struct {
	ctx         context.Context
	nc          *nats.Conn
	id          int
	appId       int
	msgChannel  chan *nats.Msg
	MessageIn   chan common.Transaction
	MessageOut  chan common.Transaction
	globalState *pbftState
}

var dummyTxn = common.Transaction{
	TxnType: "LOCAL",
	ToApp:   "",
	ToId:    "",
	FromId:  "",
	Clock:   nil,
}

func newPbftNode(ctx context.Context, nc *nats.Conn, id int, appId int, globalState *pbftState) *PbftNode {
	return &PbftNode{
		ctx:         ctx,
		nc:          nc,
		id:          id,
		appId:       appId,
		msgChannel:  make(chan *nats.Msg),
		MessageIn:   make(chan common.Transaction),
		MessageOut:  make(chan common.Transaction),
		globalState: globalState,
	}
}

func getNodeId(appId, id int) int {
	switch appId {
	case 0:
		return id
	case 1:
		return id + config.GetAppNodeCntInt(0)
	case 2:
		return id + config.GetAppNodeCntInt(0) + config.GetAppNodeCntInt(1)
	case 3:
		return id + config.GetAppNodeCntInt(0) + config.GetAppNodeCntInt(1) + config.GetAppNodeCntInt(2)
	}
	panic("no such id " + strconv.Itoa(appId))
}

func (node *PbftNode) generateGlobalGetId() func() int {
	return func() int {
		return getNodeId(node.appId, node.id)
	}
}

func (node *PbftNode) generateIsGlobalLeader(globalState *pbftState) func(int) bool {
	return func(id int) bool {
		totalNodes := config.GetAppNodeCntInt(0) + config.GetAppNodeCntInt(1) + config.GetAppNodeCntInt(2) + config.GetAppNodeCntInt(3)

		return globalState.viewNumber%totalNodes == id
	}
}

func (node *PbftNode) generateGlobalLeader(globalState *pbftState) func() bool {
	return func() bool {
		return node.generateIsGlobalLeader(globalState)(node.generateGlobalGetId()())
	}
}

func (node *PbftNode) generateSuggestedGlobalLeader() func(int) bool {
	return func(id int) bool {
		totalNodes := config.GetAppNodeCntInt(0) + config.GetAppNodeCntInt(1) + config.GetAppNodeCntInt(2) + config.GetAppNodeCntInt(3)

		return id%totalNodes == node.generateGlobalGetId()()
	}
}

func (node *PbftNode) generateGlobalBroadcast() func(common.Message) {
	return func(message common.Message) {
		message.FromApp = config.GetAppName(node.appId)
		message.FromNodeNum = node.id
		_txn := *message.Txn
		pckedMsg := packedMessage{
			Msg: message,
			Txn: _txn,
		}
		byteMessage, _ := json.Marshal(pckedMsg)
		messenger.PublishNatsMessage(node.ctx, node.nc, GLOBAL_APPLICATION, byteMessage)
	}
}

func (node *PbftNode) initTimer(state *pbftState, broadcast func(common.Message)) {
	state.timeoutTimer = time.NewTimer(TIMEOUT * time.Second)
	go func() {
		for {
			<-state.timeoutTimer.C
			broadcast(common.Message{
				MessageType: VIEW_CHANGE,
				Timestamp:   state.candidateNumber,
				FromNodeNum: node.id,
				Txn:         &dummyTxn,
			})
			state.candidateNumber++
			state.timeoutTimer.Reset(2 * TIMEOUT * time.Second)
		}
	}()
	state.timeoutTimer.Stop()
}

func (node *PbftNode) subToNatsChannels(suffix string) {
	var (
		err error
	)
	err = messenger.SubscribeToInbox(node.ctx, node.nc, GLOBAL_APPLICATION, node.msgChannel, false)
	if err != nil {
		log.WithFields(log.Fields{
			"error":       err.Error(),
			"application": GLOBAL_APPLICATION,
		}).Error("error subscribing to global nats channel")
	}
}

func (node *PbftNode) handleGlobalOut(state *pbftState) {
	txn := <-state.messageOut
	_txn := txn
	log.WithFields(log.Fields{
		"txn":   _txn,
		"id":    node.id,
		"appId": node.appId,
	}).Info("GLOBAL CONSENSUS DONE")
	node.MessageOut <- _txn
}

func (node *PbftNode) startMessageListeners(msgChan chan *nats.Msg) {
	var (
		packedMsg packedMessage
	)
	for {
		select {
		case natsMsg := <-msgChan:
			_ = json.Unmarshal(natsMsg.Data, &packedMsg)
			msg := common.Message{
				MessageType: packedMsg.Msg.MessageType,
				Timestamp:   packedMsg.Msg.Timestamp,
				FromNodeId:  packedMsg.Msg.FromNodeId,
				FromNodeNum: packedMsg.Msg.FromNodeNum,
				FromApp:     packedMsg.Msg.FromApp,
				Txn:         &packedMsg.Txn,
				Digest:      packedMsg.Msg.Digest,
				PKeySig:     packedMsg.Msg.PKeySig,
			}

			log.WithFields(log.Fields{
				"type":  packedMsg.Msg.MessageType,
				"txn":   packedMsg.Txn,
				"id":    node.id,
				"appId": node.appId,
				"inbox": GLOBAL_APPLICATION,
			}).Info("nats message received")

			node.globalState.handleMessage(
				msg,
				node.generateGlobalBroadcast(),
				node.generateGlobalLeader(node.globalState),
				node.generateSuggestedGlobalLeader(),
				node.generateGlobalGetId(),
			)

		case newTransaction := <-node.MessageIn:
			if newTransaction.TxnType != GLOBAL {
				panic("wrong transaction")
			}
			go node.generateGlobalBroadcast()(common.Message{
				MessageType: NEW_MESSAGE,
				FromNodeNum: node.id,
				FromApp:     config.GetAppName(node.appId),
				Txn:         &newTransaction,
			})
		}
	}
}
