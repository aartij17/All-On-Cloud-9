package replica

import (
	"All-On-Cloud-9/common"
	"fmt"
)

type Replica struct {
	Graph string // STUB. For now, the DAG will just be a string
	NumReplicas int
}

func (replica *Replica) HandleReceive(message *common.MessageEvent) string{
	replica.AddDepsToGraph(message)
	return replica.ExecVertices()
}

// add the dependency to the graph
func (replica *Replica) AddDepsToGraph(message *common.MessageEvent) {
	fmt.Println("add dependency to graph")
}

func (replica *Replica) ExecVertices() string{
	fmt.Println("execute every eligible vertex Vy")
	// Vy := common.Vertex{0,0}
	// if hash(vy) % num replicas = replica index then send result
	if true {
		return "success"
	} else {
		return "fail"
	}
}




