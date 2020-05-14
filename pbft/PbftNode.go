package pbft

import (
	"All-On-Cloud-9/common"
	"All-On-Cloud-9/messenger"
	"context"
	"encoding/json"
	"github.com/nats-io/nats.go"
)

type PbftNode struct {
	isPrimary  bool
	suffix string
	msgChannel chan *nats.Msg
	MessageIn  chan *nats.Msg
	MessageOut chan *nats.Msg
}

func NewPbftNode(suffix string) *PbftNode {
	return &PbftNode{
		isPrimary:false,
		suffix: suffix,
		msgChannel: make(chan *nats.Msg),
		MessageIn: make(chan *nats.Msg),
		MessageOut: make(chan *nats.Msg),
	}
}

func (node *PbftNode) startInterAppNatsListener(ctx context.Context, nc *nats.Conn, msgChan chan *nats.Msg) {
	var (
		msg *common.Message
	)
	for {
		select {
		case natsMsg := <-msgChan:
			_ = json.Unmarshal(natsMsg.Data, &msg)
		case newMessage := <-node.MessageIn:
			messenger.PublishNatsMessage(ctx, nc, _inbox(node.suffix), newMessage.Data) //todo: .Data doesn't look good
		}
	}
}