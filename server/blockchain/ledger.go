package blockchain

import (
	"All-On-Cloud-9/common"
	"fmt"
	"sync"

	log "github.com/Sirupsen/logrus"
	guuid "github.com/google/uuid"
	"github.com/hashicorp/terraform/dag"
	"github.com/hashicorp/terraform/tfdiags"
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
	IsGenesis     bool                 `json:"is_genesis"`
	BlockId       string               `json:"block_id"`
	ViewType      string               `json:"block_view_type"` //local/global block
	CryptoHash    string               `json:"crypto_hash"`
	Transaction   *common.Transaction  `json:"transaction,omitempty"`
	InitiatorNode string               `json:"initiator_node,omitempty"`
	Clock         *common.LamportClock `json:"clock"`
}

func PrintBlockchain() {
	var visits []string
	var lock sync.Mutex
	Blockchain.Walk(func(v dag.Vertex) tfdiags.Diagnostics {
		lock.Lock()
		defer lock.Unlock()
		id := v.(*Vertex).V.(*Block).BlockId
		visits = append(visits, id)
		return nil
	})

	log.Info(visits)
}

func InitBlockchain(nodeId string) *Vertex {
	id := guuid.New()
	// 1. create the Genesis block
	genesisBlock := &Block{
		BlockId:       common.LAMBDA_BLOCK,
		IsGenesis:     true,
		Transaction:   nil,
		InitiatorNode: "",
		// TODO: [Aarti]: Do we even need this?!
		Clock: &common.LamportClock{
			// TODO: [Aarti] Confirm if this is right
			PID:   fmt.Sprintf("%s-%s", nodeId, id.String()),
			Clock: common.GlobalClock,
		},
		CryptoHash: "",
	}

	// 2. Initialize the DAG
	newVertex := &Vertex{
		VertexId: common.LAMBDA_BLOCK,
		V:        dag.Vertex(genesisBlock),
	}
	Blockchain.Add(newVertex)

	return newVertex
}
