package nodes

import (
	"All-On-Cloud-9/server/blockchain"
	"context"

	log "github.com/Sirupsen/logrus"
	"github.com/hashicorp/terraform/dag"
)

func (server *Server) updateGlobalView(ctx context.Context, newBlock blockchain.Block, toVertexId string) {

}

func (server *Server) AddNewBlock(ctx context.Context, newBlock blockchain.Block, toVertexId string) {
	var (
		OK       bool
		toVertex dag.Vertex
	)
	// check if the toVertexId exists
	if toVertex, OK = server.VertexMap[toVertexId]; !OK {
		log.WithFields(log.Fields{
			"toVertexId": toVertexId,
		}).Error("toVertex not found")
		return
	}
	// to vertex found, add the new block
	// create the new vertex first
	newVertex := dag.Vertex(newBlock)
	// create an edge b/w the source and destination vertex
	edge := dag.BasicEdge(newVertex, toVertex)

	// add the edge to the DAG
	blockchain.Blockchain.Connect(edge)

	// add the vertex to the Node map
	server.VertexMap[newBlock.Transaction.TxnId] = newVertex

	// TODO: [Aarti] update the sequence global/local transaction sequence numbers

}
