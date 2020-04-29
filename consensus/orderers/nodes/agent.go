package nodes

import (
	"All-On-Cloud-9/config"
	"context"

	"github.com/nats-io/nats.go"

	log "github.com/Sirupsen/logrus"
)

var (
	NatsOrdMessage chan *nats.Msg
)

type Orderer struct {
	IsPrimary bool `json:"is_primary"`
	Id        int  `json:"agent_id"`
}

func CreateOrderer(ctx context.Context, id int) *Orderer {
	isPrimary := false
	if id == config.ORDERER1 || id == config.ORDERER2 || id == config.ORDERER3 {
		isPrimary = true
	}
	log.WithFields(log.Fields{
		"id":        id,
		"isPrimary": isPrimary,
	}).Info("initializing new orderer")
	o := &Orderer{
		Id:        id,
		IsPrimary: isPrimary,
	}
	go o.StartOrdListener(ctx)
	return o
}

func (o *Orderer) StartOrdListener(ctx context.Context) {
	var (
		msg *nats.Msg
	)
	for {
		select {
		case msg = <-NatsOrdMessage:
			_ = msg // TODO: Remove this, added to keep compiler happy
		}
	}
}
