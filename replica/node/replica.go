package replica

import (
	"All-On-Cloud-9/common"
	"fmt"
)

type Replica struct [
	Graph string // STUB. For now, the DAG will just be a string
	NumReplicas int
]

func (replica *Replica) HandleReceive(message *common.MessageEvent) {
	
}