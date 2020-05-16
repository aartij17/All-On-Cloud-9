package nodes

import (
	"All-On-Cloud-9/bpaxos"
	"All-On-Cloud-9/common"
	"All-On-Cloud-9/messenger"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/nats-io/nats.go"

	log "github.com/Sirupsen/logrus"
)

var (
	NatsOrdMessage        = make(chan *nats.Msg)
	GlobalConsensusDone   = make(chan *nats.Msg)
	ONode                 *Orderer
	numOrderMessagesRecvd = 0
	numSyncMessagesRecvd  = 0
	consensusTimeout      = 20 * time.Second
)

type Orderer struct {
	IsPrimary       bool       `json:"is_primary"`
	Id              int        `json:"agent_id"`
	NatsConn        *nats.Conn `json:"nats_connection"`
	isConsensusNode bool
	isLeader        bool
	isProposer      bool
	isReplica       bool
}

func CreateOrderer(ctx context.Context, nodeId int) error {
	var (
		isPrimary    = false
		runConsensus = false
		runLeader    = false
		runProposer  = false
		runReplica   = false
	)
	switch nodeId {
	case 0:
		isPrimary = true
		runLeader = true
	case 1:
		runProposer = true
	case 2, 3, 4:
		runConsensus = true
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
		Id:              nodeId,
		IsPrimary:       isPrimary,
		NatsConn:        nc,
		isLeader:        runLeader,
		isProposer:      runProposer,
		isReplica:       runReplica,
		isConsensusNode: runConsensus,
	}
	bpaxos.SetupBPaxos(ctx, ONode.NatsConn, runConsensus, runLeader, runProposer, runReplica)
	go ONode.StartOrdListener(ctx)
	return nil
}

// startNatsConsumers subscribes to all the NATS orderer consensus subjects
func (o *Orderer) startNatsConsumers(ctx context.Context) {
	for i := range common.NATS_ORDERER_SUBJECTS {
		_ = messenger.SubscribeToInbox(ctx, o.NatsConn, common.NATS_ORDERER_SUBJECTS[i], NatsOrdMessage)
	}
	_ = messenger.SubscribeToInbox(ctx, o.NatsConn, common.NATS_CONSENSUS_DONE_MSG, GlobalConsensusDone)
}

func (o *Orderer) initiateGlobalConsensus(ctx context.Context, natsMsg []byte) {
	messenger.PublishNatsMessage(ctx, o.NatsConn, common.NATS_CONSENSUS_INITIATE_MSG, natsMsg)
	// start a timer and wait for the global consensus to get over.
	timer := time.NewTimer(consensusTimeout)

	// change the message type to O_SYNC before publishing it to the other applications.
	var newMsg *Message
	_ = json.Unmarshal(natsMsg, &newMsg)
	newMsg.MessageType = common.O_SYNC
	natsMsg, _ = json.Marshal(newMsg)

	for {
		select {
		case <-timer.C:
			log.WithFields(log.Fields{
				"orderer_id": o.Id,
			}).Error("global consensus timeout!, no message recvd")
			return
		case <-GlobalConsensusDone:
			messenger.PublishNatsMessage(ctx, o.NatsConn, common.NATS_ORD_SYNC, natsMsg)
		}
	}
}

func (o *Orderer) StartOrdListener(ctx context.Context) {
	o.startNatsConsumers(ctx)
	var (
		natsMsg *nats.Msg
		msg     *Message
	)
	for {
		select {
		case natsMsg = <-NatsOrdMessage:
			_ = json.Unmarshal(natsMsg.Data, &msg)
			switch natsMsg.Subject {
			case common.NATS_ORD_ORDER:
				if numOrderMessagesRecvd > common.F && o.IsPrimary {
					// sufficient number of ORDER messages received, initiate global consensus
					go o.initiateGlobalConsensus(ctx, natsMsg.Data)
					numOrderMessagesRecvd = 0
				} else {
					numOrderMessagesRecvd += 1
				}
			case common.NATS_ORD_SYNC:
				// either the sync message is from a `majority` of orderer nodes, OR
				// it is from the primary orderer node, both are acceptable
				if msg.IsFromPrimary || (numSyncMessagesRecvd > common.F && o.IsPrimary) {
					// tell all the application nodes that the transaction can be added to the
					// blockchain
					messenger.PublishNatsMessage(ctx, o.NatsConn, common.NATS_ADD_TO_BC, natsMsg.Data)
					numSyncMessagesRecvd = 0
				} else {
					numSyncMessagesRecvd += 1
				}
			}
		}
	}
}
