package pbft

import (
	"All-On-Cloud-9/common"
	"All-On-Cloud-9/messenger"
	"context"
	"encoding/json"
	log "github.com/Sirupsen/logrus"
	"github.com/nats-io/nats.go"
	"reflect"
	"strconv"
	"strings"
	"time"
)

type PbftNode struct {
	ctx            context.Context
	nc             *nats.Conn
	id             int
	appId          int
	msgChannel     chan *nats.Msg
	localConsensus chan common.Transaction
	MessageIn      chan common.Transaction
	MessageOut     chan common.Transaction
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
		localConsensus: make(chan common.Transaction),
		MessageIn:      make(chan common.Transaction),
		MessageOut:     make(chan common.Transaction),
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
		//log.WithFields(log.Fields{
		//	"appId":       node.appId,
		//	"msg.Type": message.MessageType,
		//	"txn": message.Txn,
		//}).Info("broadcast locally")
		byteMessage, _ := json.Marshal(pckedMsg)
		messenger.PublishNatsMessage(node.ctx, node.nc, _inbox(node.localState.suffix), byteMessage)
	}
}

func (node *PbftNode) generateGlobalBroadcast(localState *pbftState) func(common.Message) {
	return func(message common.Message) {
		if node.generateLocalLeader(localState)() {
			_txn := *message.Txn
			_txn.Type = LOCAL_CONSENSUS + "_" + message.MessageType
			//log.WithFields(log.Fields{
			//	"txn":         _txn,
			//	"appId":       node.appId,
			//	"originalTxn": message.Txn,
			//}).Info("sending to middle consensus")
			node.MessageIn <- _txn

			var __txn common.Transaction
			for received := false; !received; {
				txn := <-node.localConsensus
				__txn = txn
				//log.WithFields(log.Fields{
				//	"txn":         __txn,
				//	"appId":       node.appId,
				//	"inbox": _inbox(node.localState.suffix),
				//}).Info("received from middle consensus")
				if !reflect.DeepEqual(__txn, _txn) {
					node.localConsensus <- __txn
				} else {
					received = true
				}
			}
			pckedMsg := packedMessage{
				Msg: message,
				Txn: __txn,
			}
			pckedMsg.Txn.Type = GLOBAL
			//pckedMsg.Txn.FromId = strconv.Itoa(node.appId)
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
		_txn := txn
		if _txn.Type == LOCAL {
			log.WithFields(log.Fields{
				"txn": _txn,
				"id": node.id,
				"appId": node.appId,
			}).Info("LOCAL CONSENSUS DONE")
			node.MessageOut <- _txn
		} else if strings.HasPrefix(_txn.Type, LOCAL_CONSENSUS) {
			if node.generateLocalLeader(node.localState)() {
				log.WithFields(log.Fields{
					"txn": _txn,
					"id": node.id,
					"appId": node.appId,
				}).Info("MIDDLE CONSENSUS DONE")
				node.localConsensus <- _txn
			}
		}
	}
}

func (node *PbftNode) handleGlobalOut(state *pbftState) {
	txn := <-state.messageOut
	_txn := txn
	log.WithFields(log.Fields{
		"txn": _txn,
		"id": node.id,
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
				Txn:         &packedMsg.Txn,
				Digest:      packedMsg.Msg.Digest,
				PKeySig:     packedMsg.Msg.PKeySig,
			}

			//log.WithFields(log.Fields{
			//	"type":        packedMsg.Msg.MessageType,
			//	"txn":         packedMsg.Txn,
			//	"id":          node.id,
			//	"appId":       node.appId,
			//	"inbox": _inbox(node.localState.suffix),
			//}).Info("nats message received")
			//msg.Txn.FromId = ""

			if msg.Txn.Type == LOCAL || strings.HasPrefix(msg.Txn.Type, LOCAL_CONSENSUS) {
				node.localState.handleMessage(
					msg,
					node.generateLocalBroadcast(),
					node.generateLocalLeader(node.localState),
					node.generateSuggestedLocalLeader(node.localState),
					node.generateLocalGetId(),
				)
			} else if msg.Txn.Type == GLOBAL && (node.generateLocalLeader(node.localState)() || msg.MessageType == COMMITED) {
				node.globalState.handleMessage(
					msg,
					node.generateGlobalBroadcast(node.localState),
					node.generateGlobalLeader(node.globalState, node.localState),
					node.generateSuggestedGlobalLeader(node.globalState, node.localState),
					node.generateGlobalGetId(),
				)
			}

		case newMessage := <-node.MessageIn:
			if newMessage.Type == LOCAL || strings.HasPrefix(newMessage.Type, LOCAL_CONSENSUS) {
				go node.generateLocalBroadcast()(common.Message{
					MessageType: NEW_MESSAGE,
					FromNodeNum: node.id,
					Txn:         &newMessage,
				})
			} else {
				go node.generateGlobalBroadcast(node.localState)(common.Message{
					MessageType: NEW_MESSAGE,
					FromNodeNum: node.id,
					Txn:         &newMessage,
				})
			}
		}
	}
}
