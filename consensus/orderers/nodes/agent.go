package nodes

import (
	"All-On-Cloud-9/common"
	"All-On-Cloud-9/messenger"
	"context"
	"encoding/json"
	"fmt"

	"github.com/nats-io/nats.go"

	log "github.com/Sirupsen/logrus"
)

var (
	NatsOrdMessage = make(chan *nats.Msg)
	ONode          *Orderer
)

type Orderer struct {
	IsPrimary bool       `json:"is_primary"`
	Id        int        `json:"agent_id"`
	NatsConn  *nats.Conn `json:"nats_connection"`
}

//ctx, nodeId, appName, nodeIdNum
func CreateOrderer(ctx context.Context, nodeId int) error {
	isPrimary := false
	if nodeId == 1 {
		isPrimary = true
	}
	log.WithFields(log.Fields{
		"id":        nodeId,
		"isPrimary": isPrimary,
	}).Info("initializing new orderer")

	nc, err := messenger.NatsConnect(ctx)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Error("error connecting to nats")
		return fmt.Errorf("error connecting to NATS")
	}

	ONode = &Orderer{
		Id:        nodeId,
		IsPrimary: isPrimary,
		NatsConn:  nc,
	}
	go ONode.StartOrdListener(ctx)
	return nil
}

// startNatsConsumers subscribes to all the NATS orderer consensus subjects
func (o *Orderer) startNatsConsumers(ctx context.Context) {
	for i := range common.NATS_ORDERER_SUBJECTS {
		_ = messenger.SubscribeToInbox(ctx, o.NatsConn, common.NATS_ORDERER_SUBJECTS[i], NatsOrdMessage)
	}
}

func (o *Orderer) StartOrdListener(ctx context.Context) {
	o.startNatsConsumers(ctx)
	var (
		natsMsg *nats.Msg
		msg     *Message
		_       error
	)
	for {
		select {
		case natsMsg = <-NatsOrdMessage:
			_ = json.Unmarshal(natsMsg.Data, &msg)
			switch natsMsg.Subject {
			case common.NATS_ORD_ORDER:
				continue
			case common.NATS_ORD_SYNC:
				continue
			}
		}
	}
}
