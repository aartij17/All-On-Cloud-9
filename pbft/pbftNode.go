package pbft

import (
	"All-On-Cloud-9/common"
	"All-On-Cloud-9/messenger"
	"context"
	"encoding/json"
	log "github.com/Sirupsen/logrus"
	"github.com/nats-io/nats.go"
	"time"
)

type PbftNode struct {
	ctx         context.Context
	nc          *nats.Conn
	id          int
	appId       int
	msgChannel  chan *nats.Msg
	MessageIn   chan *common.Transaction
	MessageOut  chan *common.Transaction
	localState  *pbftState
	globalState *pbftState
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
		ctx:         ctx,
		nc:          nc,
		id:          id,
		appId:       appId,
		msgChannel:  make(chan *nats.Msg),
		MessageIn:   make(chan *common.Transaction),
		MessageOut:  make(chan *common.Transaction),
		localState:  localState,
		globalState: globalState,
	}
}

func (node *PbftNode) generateGetId() func() int {
	return func() int {
		return node.id
	}
}

func (node *PbftNode) generateLocalLeader(state *pbftState) func() bool {
	return func() bool {
		return state.viewNumber%state.totalNodes == node.id && node.id != 0
	}
}

func (node *PbftNode) generateSuggestedLocalLeader(state *pbftState) func(int) bool {
	return func(id int) bool {
		return id%state.totalNodes == node.id
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
				"application": APPLICATION,
			}).Info("nats message received")
			if msg.Txn.Type == LOCAL {
				txn := node.localState.handleMessage(
					msg,
					node.generateLocalBroadcast(),
					node.generateLocalLeader(node.localState),
					node.generateSuggestedLocalLeader(node.localState),
					node.generateGetId(),
				)

				if txn != nil {
					log.WithField("txn", txn).Info("SUCCESS")
					node.MessageOut <- txn
				}
			} else if msg.Txn.Type == GLOBAL {
				//node.handleGlobalMessage
			}

		case newMessage := <-node.MessageIn:
			node.generateLocalBroadcast()(common.Message{
				MessageType: NEW_MESSAGE,
				FromNodeNum: node.id,
				Txn:         newMessage,
			})
		}
	}
}
