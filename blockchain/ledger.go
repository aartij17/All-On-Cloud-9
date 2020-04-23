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
	Digit       int                 `json:"digit"`
	Transaction *common.Transaction `json:"transaction"`
}
