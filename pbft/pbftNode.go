package pbft

import (
	"All-On-Cloud-9/common"
	"All-On-Cloud-9/messenger"
	"context"
	"encoding/json"
	log "github.com/Sirupsen/logrus"
	"github.com/nats-io/nats.go"
)

type reducedMessage struct {
	messageType string
	Txn *common.Transaction
}

type PbftNode struct {
	ctx              context.Context
	nc               *nats.Conn
	viewNumber       int
	id               int
	failureTolerance int
	totalNodes       int
	counter			 map[reducedMessage]int
	suffix           string
	msgChannel       chan *nats.Msg
	MessageIn        chan *common.Transaction
	MessageOut       chan *common.Transaction
	localLog         []common.Message
}

func newPbftNode(ctx context.Context, nc *nats.Conn, suffix string, failureTolerance int, totalNodes int, id int) *PbftNode {
	return &PbftNode{
		ctx:              ctx,
		nc:               nc,
		viewNumber:       0,
		id:               id,
		failureTolerance: failureTolerance,
		totalNodes:       totalNodes,
		counter:          make(map[reducedMessage]int),
		suffix:           suffix,
		msgChannel:       make(chan *nats.Msg),
		MessageIn:        make(chan *common.Transaction),
		MessageOut:       make(chan *common.Transaction),
		localLog:         make([]common.Message, 0),
	}
}

func (node *PbftNode) broadcast(message common.Message) {
	byteMessage, _ := json.Marshal(message)
	messenger.PublishNatsMessage(node.ctx, node.nc, _inbox(node.suffix), byteMessage)
}

func (node *PbftNode) getLeaderId() int {
	return node.viewNumber % node.totalNodes
}

func (node *PbftNode) isLeader() bool {
	return node.getLeaderId() == node.id
}

func (node *PbftNode) handleLocalMessage(message common.Message) {
	currentTimestamp := -1

	switch message.MessageType {
	case NEW_MESSAGE:
		if node.isLeader() {
			currentTimestamp++
			node.broadcast(common.Message{
				MessageType: PRE_PREPARE,
				Timestamp:   currentTimestamp,
				FromNodeNum: node.id,
				Txn:         message.Txn,
			})
		}
	case PRE_PREPARE:
		node.broadcast(common.Message{
			MessageType: PREPARE,
			Timestamp:   message.Timestamp,
			FromNodeNum: node.id,
			Txn:         message.Txn,
		})
	case PREPARE:
		reduced := reducedMessage{
			messageType: PREPARE,
			Txn:         message.Txn,
		}

		node.counter[reduced]++
		if node.counter[reduced] >= 2*node.failureTolerance+1 {
			node.broadcast(common.Message{
				MessageType: COMMIT,
				Timestamp:   message.Timestamp,
				FromNodeNum: node.id,
				Txn:         message.Txn,
			})
		}
	case COMMIT:
		reduced := reducedMessage{
			messageType: COMMIT,
			Txn:         message.Txn,
		}

		node.counter[reduced]++
		if node.counter[reduced] >= 2*node.failureTolerance+1 {
			node.broadcast(common.Message{
				MessageType: COMMITED,
				Timestamp:   message.Timestamp,
				FromNodeNum: node.id,
				Txn:         message.Txn,
			})
		}
	case COMMITED:
		reduced := reducedMessage{
			messageType: COMMITED,
			Txn:         message.Txn,
		}

		node.counter[reduced]++
		if node.counter[reduced] >= node.failureTolerance + 1 {
			node.MessageOut <- message.Txn
		}
	}
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
		msg common.Message
	)
	for {
		select {
		case natsMsg := <-msgChan:
			_ = json.Unmarshal(natsMsg.Data, &msg)
			log.WithFields(log.Fields{
				"nats message":      msg,
				"application": APPLICATION,
			}).Info("nats message received")
			if msg.Txn.Type == "LOCAL" {
				node.handleLocalMessage(msg)
			} else if msg.Txn.Type == "GLOBAL" {
				//node.handleGlobalMessage
			}

		case newMessage := <-node.MessageIn:
			node.broadcast(common.Message{
				MessageType: NEW_MESSAGE,
				FromNodeNum: node.id,
				Txn:         newMessage,
			})
		}
	}
}