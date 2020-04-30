package nodes

import (
	"All-On-Cloud-9/server/blockchain"
	"context"

	"github.com/hashicorp/terraform/dag"
)

type Node struct {
	Id                  string                `json:"server_id"`
	AppName             string                `json:"appname"`
	NumId               int                   `json:"numeric_id"`
	IsAgent             bool                  `json:"is_agent"`
	VertexMap           map[string]dag.Vertex `json:"vertex_map"`
	CurrentLocalTxnSeq  int                   `json:"current_local_txn_seq"`
	CurrentGlobalTxnSeq int                   `json:"current_global_txn_seq"`
}

func StartServer(ctx context.Context, nodeId string, appName string, id int) *Node {
	rootNode := &Node{
		Id:      nodeId,
		AppName: appName,
		NumId:   id,
		IsAgent: false,
	}

	// initialize the public blockchain
	blockchain.InitBlockchain(id)
	return rootNode
}
