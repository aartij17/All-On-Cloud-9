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

type packedMessage struct {
	Msg common.Message     `json:"msg"`
	Txn common.Transaction `json:"txn"`
}

type reducedMessage struct {
	messageType string
	Txn         *common.Transaction
}

type PbftNode struct {
	ctx              context.Context
	nc               *nats.Conn
	viewNumber       int
	id               int
	failureTolerance int
	totalNodes       int
	counter          map[reducedMessage]int
	suffix           string
	msgChannel       chan *nats.Msg
	MessageIn        chan *common.Transaction
	MessageOut       chan *common.Transaction
	localLog         []common.Message

	currentTimestamp  int
	viewChangeCounter int
	timerIsRunning    bool
	candidateNumber   int
	timeoutTimer      *time.Timer
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
	pckedMsg := packedMessage{
		Msg: message,
		Txn: *message.Txn,
	}
	byteMessage, _ := json.Marshal(pckedMsg)
	messenger.PublishNatsMessage(node.ctx, node.nc, _inbox(node.suffix), byteMessage)
}

func (node *PbftNode) getLeaderId() int {
	return node.viewNumber % node.totalNodes
}

func (node *PbftNode) isLeader() bool {
	return node.getLeaderId() == node.id
}

func (node *PbftNode) handleLocalMessage(message common.Message) {
	switch message.MessageType {
	case NEW_VIEW:
		node.viewChangeCounter = 0
		node.timerIsRunning = false
		node.candidateNumber = node.viewNumber + 1
		node.timeoutTimer.Stop()
		node.viewNumber = message.Timestamp
	case VIEW_CHANGE:
		if message.Timestamp%node.totalNodes == node.id {
			node.viewChangeCounter++
			println("inside VIEW_CHANGE", node.id, node.viewChangeCounter, 2*node.failureTolerance+1)
			if node.viewChangeCounter == 2*node.failureTolerance+1 {
				println("broadcasting NEW_VIEW")
				node.broadcast(common.Message{
					MessageType: NEW_VIEW,
					Timestamp:   message.Timestamp,
					FromNodeNum: node.id,
					Txn:         &dummyTxn,
				})
			}
		}
	case NEW_MESSAGE:
		if node.isLeader() {
			node.currentTimestamp++
			node.broadcast(common.Message{
				MessageType: PRE_PREPARE,
				Timestamp:   node.currentTimestamp,
				FromNodeNum: node.id,
				Txn:         message.Txn,
			})
		} else {
			if !node.timerIsRunning {
				node.timeoutTimer.Reset(TIMEOUT * time.Second)
				node.timerIsRunning = true
			}
		}
	case PRE_PREPARE:
		if message.FromNodeNum == node.getLeaderId() {
			node.timeoutTimer.Stop()
			node.timerIsRunning = false
			node.broadcast(common.Message{
				MessageType: PREPARE,
				Timestamp:   message.Timestamp,
				FromNodeNum: node.id,
				Txn:         message.Txn,
			})
		}
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
		if node.counter[reduced] >= node.failureTolerance+1 {
			node.currentTimestamp = message.Timestamp
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
	node.currentTimestamp = -1
	node.viewChangeCounter = 0
	node.timerIsRunning = false
	node.candidateNumber = node.viewNumber + 1
	node.timeoutTimer = time.NewTimer(TIMEOUT * time.Second)
	go func() {
		for {
			<-node.timeoutTimer.C
			node.broadcast(common.Message{
				MessageType: VIEW_CHANGE,
				Timestamp:   node.candidateNumber,
				FromNodeNum: node.id,
				Txn:         &dummyTxn,
			})
			node.candidateNumber++
			node.timeoutTimer.Reset(2 * TIMEOUT * time.Second)
		}
	}()
	node.timeoutTimer.Stop()

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
				//"nats message": packedMsg,
				"msg": packedMsg.Msg,
				"txn": packedMsg.Txn,
				"id":  node.id,
				//"type":         msg.Txn.Type,
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
