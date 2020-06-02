package blockchain

import (
	"All-On-Cloud-9/common"
	"fmt"
	"sync"

	guuid "github.com/google/uuid"
	"github.com/hashicorp/terraform/dag"
	"github.com/hashicorp/terraform/tfdiags"
	log "github.com/sirupsen/logrus"
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

type GraphPrint struct {
	visits        []string
	outgoingEdges []string
	incomingEdges []string
}

func PrintBlockchain() {
	var gpArr []GraphPrint
	var lock sync.Mutex
	Blockchain.Walk(func(v dag.Vertex) tfdiags.Diagnostics {
		gp := GraphPrint{
			visits:        make([]string, 0),
			outgoingEdges: make([]string, 0),
			incomingEdges: make([]string, 0),
		}
		lock.Lock()
		defer lock.Unlock()
		id := v.(*Vertex).V.(*Block).BlockId
		outgoingEdges := Blockchain.EdgesFrom(v)
		for e := range outgoingEdges {
			gp.outgoingEdges = append(gp.outgoingEdges, outgoingEdges[e].Target().(*Vertex).V.(*Block).BlockId)
		}
		incomingEdges := Blockchain.EdgesTo(v)
		for e := range incomingEdges {
			gp.incomingEdges = append(gp.incomingEdges, incomingEdges[e].Source().(*Vertex).V.(*Block).BlockId)
		}
		gp.visits = append(gp.visits, id)
		gpArr = append(gpArr, gp)
		return nil
	})
	log.Warn("**********************************************************************************************")
	log.Warn("[Node], [Outgoing edges], [Incoming Edges]")
	for gp := range gpArr {
		log.Warn(gpArr[gp])
	}
	log.Warn("**********************************************************************************************")
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
