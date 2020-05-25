package bpaxos

import (
	consensus "All-On-Cloud-9/bpaxos/consensus/node"
	leadernode "All-On-Cloud-9/bpaxos/leader/node"
	proposer "All-On-Cloud-9/bpaxos/proposer/node"
	replica "All-On-Cloud-9/bpaxos/replica/node"
	"context"

	log "github.com/Sirupsen/logrus"

	"github.com/nats-io/nats.go"
)

var (
	// Keep a running counter so that
	// all leaders will have a unique index
	leader_count = 0
)

func SetupBPaxos(ctx context.Context, nc *nats.Conn, runConsensus bool,
	runLeader bool, runProposer bool, runReplica bool) {

	log.WithFields(log.Fields{
		"isConsensus": runConsensus,
		"isLeader":    runLeader,
		"isProposer":  runProposer,
		"isReplica":   runReplica,
	}).Info("bootstrapping BPAXOS roles")

	if runLeader {
		go leadernode.StartLeader(ctx, nc, leader_count)
		leader_count += 1
	}

	if runProposer {
		go proposer.StartProposer(ctx, nc)
	}

	if runConsensus {
		go consensus.StartConsensus(ctx, nc)
	}

	if runReplica {
		go replica.StartReplica(ctx, nc)
	}
}