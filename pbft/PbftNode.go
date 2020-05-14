package pbft

import (
	"All-On-Cloud-9/common"
	"context"
	"encoding/json"
	"github.com/nats-io/nats.go"
)

type PbftNode struct {
	isPrimary bool
	MsgChannel    chan *nats.Msg
}

func NewPbftNode() *PbftNode {
	return &PbftNode{isPrimary:false}
}

func (node *PbftNode) startInterAppNatsListener(ctx context.Context, msgChan chan *nats.Msg) {
	var (
		msg *common.Message
	)
	for {
		select {
		case natsMsg := <-msgChan:
			_ = json.Unmarshal(natsMsg.Data, &msg)

		}
	}
}