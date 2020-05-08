package application

import (
	"All-On-Cloud-9/common"
	"All-On-Cloud-9/config"
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
			switch msg.FromApp {
			case config.APP_MANUFACTURER:
				manufacturer.processTxn(ctx, msg)
			case config.APP_SUPPLIER:
				supplier.processTxn(ctx, msg)
			case config.APP_CARRIER:
				carrier.processTxn(ctx, msg)
			case config.APP_BUYER:
				buyer.processTxn(ctx, msg)
			}
		}
	}
}
