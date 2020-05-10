package application

import (
	"All-On-Cloud-9/common"
	"All-On-Cloud-9/config"
	"context"
	"encoding/json"
	"net/http"

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
			switch msg.ToApp {
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

func startClient(ctx context.Context, addr string, port string, handler func(http.ResponseWriter, *http.Request)) error {
	http.HandleFunc(addr, handler)
	err := http.ListenAndServe(":" + port, nil)
	return err
}

