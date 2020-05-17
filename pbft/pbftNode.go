package pbft

import (
	"All-On-Cloud-9/common"
	"All-On-Cloud-9/messenger"
	"context"
	"encoding/json"
	log "github.com/Sirupsen/logrus"
	"github.com/nats-io/nats.go"
	"reflect"
	"time"
)

type PbftNode struct {
	ctx            context.Context
	nc             *nats.Conn
	id             int
	appId          int
	msgChannel     chan *nats.Msg
	localConsensus chan *common.Transaction
	MessageIn      chan *common.Transaction
	MessageOut     chan *common.Transaction
	localState     *pbftState
	globalState    *pbftState
}

var dummyTxn = common.Transaction{
	LocalXNum:  "0",
	GlobalXNum: "0",
	Type:       "LOCAL",
	TxnId:      "",
	ToId:       "",
	FromId:     "",
	CryptoHash: "",
	TxnType:    "",
	Clock:      nil,
}

func newPbftNode(ctx context.Context, nc *nats.Conn, id int, appId int, localState *pbftState, globalState *pbftState) *PbftNode {
	return &PbftNode{
		ctx:            ctx,
		nc:             nc,
		id:             id,
		appId:          appId,
		msgChannel:     make(chan *nats.Msg),
		localConsensus: make(chan *common.Transaction),
		MessageIn:      make(chan *common.Transaction),
		MessageOut:     make(chan *common.Transaction),
		localState:     localState,
		globalState:    globalState,
	}
}

func (node *PbftNode) generateLocalGetId() func() int {
	return func() int {
		return node.id
	}
}

func (node *PbftNode) generateGlobalGetId() func() int {
	return func() int {
		return -node.id
	}
}

func (node *PbftNode) generateLocalLeader(state *pbftState) func() bool {
	return func() bool {
		return state.viewNumber%state.totalNodes == node.id
	}
}

func (node *PbftNode) generateGlobalLeader(globalState *pbftState, localState *pbftState) func() bool {
	return func() bool {
		return globalState.viewNumber%globalState.totalNodes == node.appId && node.generateLocalLeader(localState)()
	}
}

func (node *PbftNode) generateSuggestedLocalLeader(state *pbftState) func(int) bool {
	return func(id int) bool {
		return id%state.totalNodes == node.id
	}
}

func (node *PbftNode) generateSuggestedGlobalLeader(globalState *pbftState, localState *pbftState) func(int) bool {
	return func(id int) bool {
		return id%globalState.totalNodes == node.appId && node.generateLocalLeader(localState)()
	}
}

func (node *PbftNode) generateLocalBroadcast() func(common.Message) {
	return func(message common.Message) {
		pckedMsg := packedMessage{
			Msg: message,
			Txn: *message.Txn,
		}
		byteMessage, _ := json.Marshal(pckedMsg)
		messenger.PublishNatsMessage(node.ctx, node.nc, _inbox(node.localState.suffix), byteMessage)
	}
}

func (node *PbftNode) generateGlobalBroadcast(localState *pbftState) func(common.Message) {
	return func(message common.Message) {
		if node.generateLocalLeader(localState)() {
			message.Txn.Type = LOCAL_CONSENSUS
			node.generateLocalBroadcast()(message)

			txn := <-node.localConsensus
			//log.WithField("txn", txn).Info("BROADCAST RECEIVED TXN")
			if !reflect.DeepEqual(txn, message.Txn) {
				node.localConsensus <- txn
				return
			}
			message.Txn.Type = GLOBAL
			pckedMsg := packedMessage{
				Msg: message,
				Txn: *message.Txn,
			}
			byteMessage, _ := json.Marshal(pckedMsg)
			messenger.PublishNatsMessage(node.ctx, node.nc, GLOBAL_APPLICATION, byteMessage)
		}
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
	err = messenger.SubscribeToInbox(node.ctx, node.nc, _inbox(suffix), node.msgChannel)
	if err != nil {
		log.WithFields(log.Fields{
			"error":       err.Error(),
			"application": APPLICATION,
			"inbox":       _inbox(suffix),
		}).Error("error subscribing to local nats channel")
	}
	err = messenger.SubscribeToInbox(node.ctx, node.nc, GLOBAL_APPLICATION, node.msgChannel)
	if err != nil {
		log.WithFields(log.Fields{
			"error":       err.Error(),
			"application": GLOBAL_APPLICATION,
		}).Error("error subscribing to global nats channel")
	}
}

func (node *PbftNode) handleLocalOut(state *pbftState) {
	for {
		txn := <-state.messageOut
		if txn.Type == LOCAL {
			//log.WithField("txn", txn).Info("SUCCESS")
			node.MessageOut <- txn
		} else if txn.Type == LOCAL_CONSENSUS {
			//log.WithField("txn", txn).Info("CONSENSUS OUT")
			node.localConsensus <- txn
		}
	}
}

func (node *PbftNode) handleGlobalOut(state *pbftState) {
	txn := <-state.messageOut
	node.MessageOut <- txn
}

func (node *PbftNode) startMessageListeners(msgChan chan *nats.Msg) {
	var (
		packedMsg packedMessage
		msg       common.Message
	)
	for {
		select {
		case natsMsg := <-msgChan:
			_ = json.Unmarshal(natsMsg.Data, &packedMsg)
			msg = common.Message{
				MessageType: packedMsg.Msg.MessageType,
				Timestamp:   packedMsg.Msg.Timestamp,
				FromNodeId:  packedMsg.Msg.FromNodeId,
				FromNodeNum: packedMsg.Msg.FromNodeNum,
				Txn:         &packedMsg.Txn,
				Digest:      packedMsg.Msg.Digest,
				PKeySig:     packedMsg.Msg.PKeySig,
			}

			log.WithFields(log.Fields{
				"msg":         packedMsg.Msg,
				"txn":         packedMsg.Txn,
				"id":          node.id,
				"appId":       node.appId,
				"application": APPLICATION,
			}).Info("nats message received")
			if msg.Txn.Type == LOCAL || msg.Txn.Type == LOCAL_CONSENSUS {
				node.localState.handleMessage(
					msg,
					node.generateLocalBroadcast(),
					node.generateLocalLeader(node.localState),
					node.generateSuggestedLocalLeader(node.localState),
					node.generateLocalGetId(),
				)
			} else if msg.Txn.Type == GLOBAL {
				node.globalState.handleMessage(
					msg,
					node.generateGlobalBroadcast(node.localState),
					node.generateGlobalLeader(node.globalState, node.localState),
					node.generateSuggestedGlobalLeader(node.globalState, node.localState),
					node.generateGlobalGetId(),
				)
			}

		case newMessage := <-node.MessageIn:
			if newMessage.Type == LOCAL {
				go node.generateLocalBroadcast()(common.Message{
					MessageType: NEW_MESSAGE,
					FromNodeNum: node.id,
					Txn:         newMessage,
				})
			} else {
				go node.generateGlobalBroadcast(node.localState)(common.Message{
					MessageType: NEW_MESSAGE,
					FromNodeNum: node.id,
					Txn:         newMessage,
				})
			}
		}
	}
}
