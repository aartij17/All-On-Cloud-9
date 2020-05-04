package bpaxos

import (
	"All-On-Cloud-9/bpaxos/leader/node"
	"All-On-Cloud-9/bpaxos/proposer/node"
	"All-On-Cloud-9/bpaxos/dependency/node"
	"All-On-Cloud-9/bpaxos/consensus/node"
	"All-On-Cloud-9/bpaxos/replica/node"
)

var (
	// Keep a running counter so that 
	// all leaders will have a unique index
	leader_count = 0
)

// // Create a UDP Server listening on a port and return the port
// func CreateUDPSocket() (int, net.Listener) {
//     // Keep trying ports until either you exhaust all ports 
//     // or one is available
//     // If no ports are available, return 0
//     var port int
//     var l net.Listener
//     var err error
//     for port = 8000; port <= 65535; port++ {
//         l, err = net.Listen("udp", fmt.Sprintf("127.0.0.1:%d", port))
//         if err == nil {
//             fmt.Println(fmt.Sprintf("Port %d selected", port) )
//             return port, l
//         }
//     }

//     return 0, nil
// }

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


