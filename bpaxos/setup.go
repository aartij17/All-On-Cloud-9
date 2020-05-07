package bpaxos

import (
	"All-On-Cloud-9/bpaxos/consensus/node"
	"All-On-Cloud-9/bpaxos/leader/node"
	"All-On-Cloud-9/bpaxos/proposer/node"
	"All-On-Cloud-9/bpaxos/replica/node"
	"context"
	"github.com/nats-io/nats.go"
)

var (
	// Keep a running counter so that
	// all leaders will have a unique index
	leader_count = 0
)

func SetupBPaxos(ctx context.Context, nc *nats.Conn, runConsensus bool, runLeader bool, runProposer bool, runReplica bool) {

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
