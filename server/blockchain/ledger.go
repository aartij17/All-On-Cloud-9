package blockchain

import (
	"All-On-Cloud-9/common"

	"github.com/hashicorp/terraform/dag"
)

/**
Usage:
type Block struct {
	Digit int `json:"digit"`
}
var g dag.AcyclicGraph
g.Add(Block{Digit:1})

https://github.com/hashicorp/terraform/blob/master/dag/dag_test.go
*/
var (
	GlobalSeqNumber  int
	LocalSeqNumber   int
	Blockchain       dag.AcyclicGraph
	GlobalBlockchain dag.AcyclicGraph
)

func InitBlockchain(nodeId int) {
	// 1. create the Genesis block
	genesisBlock := &Block{
		IsGenesis:     true,
		Transaction:   nil,
		InitiatorNode: "",
		Clock: &common.LamportClock{
			// TODO: [Aarti] Confirm if this is right
			PID:   nodeId,
			Clock: 0,
		},
	}
	// 2. Initialize the DAG
	Blockchain.Add(dag.Vertex(genesisBlock))
}

type Block struct {
	IsGenesis     bool                 `json:"is_genesis"`
	Transaction   *common.Transaction  `json:"transaction,omitempty"`
	InitiatorNode string               `json:"initiator_node,omitempty"`
	Clock         *common.LamportClock `json:"clock"`
}
