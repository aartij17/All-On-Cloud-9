package bpaxos

import (
	"All-On-Cloud-9/bpaxos/consensus/node"
	"All-On-Cloud-9/bpaxos/dependency/node"
	"All-On-Cloud-9/bpaxos/leader/node"
	"All-On-Cloud-9/bpaxos/proposer/node"
	"All-On-Cloud-9/bpaxos/replica/node"
	"context"
)

var (
	// Keep a running counter so that
	// all leaders will have a unique index
	leader_count = 0
)

func SetupBPaxos(ctx context.Context, isPrimary bool) {

	if isPrimary {
		go leadernode.StartLeader(ctx,leader_count)
		go proposer.StartProposer(ctx)
		leader_count += 1
	}

	go depsnode.StartDependencyService(ctx)
	go consensus.StartConsensus(ctx)
	go replica.StartReplica(ctx)
}
