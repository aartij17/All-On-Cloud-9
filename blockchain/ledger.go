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

var publicBC dag.AcyclicGraph

func InitPublicBlockchain() {
	// 1. create the Genesis block
	// 2. Initialize the DAG
}

type Block struct {
	LocalSeqNum  int                  `json:"local_seq_num"`
	GlobalSeqNum int                  `json:"global_seq_num"`
	Clock        *common.LamportClock `json:"clock"`
	Transaction  *common.Transaction  `json:"transaction"`
}
