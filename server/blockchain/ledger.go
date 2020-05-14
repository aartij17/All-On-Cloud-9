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
	GlobalSeqNumber = 0
	LocalSeqNumber  = 0
	Blockchain      dag.AcyclicGraph
	//GlobalBlockchain dag.AcyclicGraph
)

type Vertex struct {
	VertexId string     `json:"vertex_id"`
	V        dag.Vertex `json:"vertex"`
}

type Block struct {
	IsGenesis bool   `json:"is_genesis"`
	BlockId   string `json:"block_id"`
	//LocalBlockNum  string               `json:"local_block_num,omitempty"`
	//GlobalBlockNum string               `json:"global_block_num,omitempty"`
	ViewType      string               `json:"block_view_type"` //local/global block
	CryptoHash    string               `json:"crypto_hash"`
	Transaction   *common.Transaction  `json:"transaction,omitempty"`
	InitiatorNode string               `json:"initiator_node,omitempty"`
	Clock         *common.LamportClock `json:"clock"`
}

func InitBlockchain(nodeId int) *Vertex {
	// 1. create the Genesis block
	genesisBlock := &Block{
		BlockId:       common.LAMBDA_BLOCK,
		IsGenesis:     true,
		Transaction:   nil,
		InitiatorNode: "",
		// TODO: [Aarti]: Do we even need this?!
		Clock: &common.LamportClock{
			// TODO: [Aarti] Confirm if this is right
			PID:   nodeId,
			Clock: LocalSeqNumber,
		},
		CryptoHash: "",
	}
	//// increment the count of both the global and local sequence number since the genesis block is
	//// visible in both local and global view
	//LocalSeqNumber += 1
	//GlobalSeqNumber += 1

	// 2. Initialize the DAG
	newVertex := &Vertex{
		VertexId: common.LAMBDA_BLOCK,
		V:        dag.Vertex(genesisBlock),
	}
	Blockchain.Add(newVertex)

	return newVertex
}
