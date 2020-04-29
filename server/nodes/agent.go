package nodes

import (
	"All-On-Cloud-9/blockchain"
	"context"

	"github.com/hashicorp/terraform/dag"
)

var PrivateBC dag.AcyclicGraph

type Node struct {
	Id string `json:"server_id"`
	// IsAgent: every node in the system
	IsAgent   bool `json:"is_agent"`
	PrivateBC dag.AcyclicGraph
}

func GetPrivateBC() dag.AcyclicGraph {
	return PrivateBC
}

func privateBlockChainInit() {

}

func StartServer(ctx context.Context, id string) *Node {
	rootNode := &Node{
		Id:        id,
		IsAgent:   false,
		PrivateBC: PrivateBC,
	}
	// initialize the public and private blockchain
	privateBlockChainInit()

	// initialize the public blockchain
	blockchain.InitPublicBlockchain()
	return rootNode
}
