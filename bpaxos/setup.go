package bpaxos

import (
	"All-On-Cloud-9/bpaxos/consensus/node"
	"All-On-Cloud-9/bpaxos/dependency/node"
	"All-On-Cloud-9/bpaxos/leader/node"
	"All-On-Cloud-9/bpaxos/proposer/node"
	"All-On-Cloud-9/bpaxos/replica/node"
)

var (
	// Keep a running counter so that
	// all leaders will have a unique index
	leader_count = 0
)

func SetupBPaxos(isPrimary bool) {

	if isPrimary {
		go leadernode.StartLeader(leader_count)
		go proposer.StartProposer()
		leader_count += 1
	}

	go depsnode.StartDependencyService()
	go consensus.StartConsensus()
	go replica.StartReplica()
}
