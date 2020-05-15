package application

import (
	"All-On-Cloud-9/common"
	"context"
	"encoding/json"

	"github.com/nats-io/nats.go"
)

func startInterAppNatsListener(ctx context.Context, msgChan chan *nats.Msg) {
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
