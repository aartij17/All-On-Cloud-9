package blockchain

import (
	"All-On-Cloud-9/common"
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


func InitPublicBlockchain() {
	// 1. create the Genesis block
	// 2. Initialize the DAG
}

type Block struct {
	Type          string               `json:"transaction_type"` // local/global transaction type
	LocalXNum     string               `json:"local_transaction_number,omitempty"`
	GlobalXNum    string               `json:"global_transaction_number,omitempty"`
	Clock         *common.LamportClock `json:"clock"`
	Transaction   *common.Transaction  `json:"transaction"`
	ToEdges       []string             `json:"to_edges"`
	FromEdges     []string             `json:"from_edges"`
	VisibleToApps []string             `json:"visible_to_apps"`
}
