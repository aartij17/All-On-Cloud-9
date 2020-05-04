package replica

import (
	"All-On-Cloud-9/common"
	"encoding/json"
	"fmt"
	"github.com/nats-io/nats.go"
)

type Replica struct {
	Graph       string // STUB. For now, the DAG will just be a string
	NumReplicas int
}

func (replica *Replica) HandleReceive(message *common.MessageEvent) string {
	replica.AddDepsToGraph(message)
	return replica.ExecVertices()
}

// add the dependency to the graph
func (replica *Replica) AddDepsToGraph(message *common.MessageEvent) {
	fmt.Println("add dependency to graph")
}

func (replica *Replica) ExecVertices() string {
	fmt.Println("execute every eligible vertex Vy")
	// Vy := common.Vertex{0,0}
	// if hash(vy) % num replicas = replica index then send result
	if true {
		return "success"
	} else {
		return "fail"
	}
}

func StartReplica() {
	rep := Replica{}
	go func(rep *Replica) {
		socket := common.Socket{}
		_ = socket.Connect(nats.DefaultURL)
		socket.Subscribe(common.ProposerToReplica, func(m *nats.Msg) {
			fmt.Println("Received proposer to replica")
			data := common.MessageEvent{}
			json.Unmarshal(m.Data, &data)
			newMessage := rep.HandleReceive(&data)
			sentMessage, err := json.Marshal(&newMessage)
			// Respond back to the client
			if err == nil {
				fmt.Println("leader can publish a message to deps")
				socket.Publish(common.NATS_CONSENSUS_DONE, sentMessage)
			} else {
				fmt.Println("json marshal failed")
				fmt.Println(err.Error())
			}

		})
	}(&rep)
	common.HandleInterrupt()
}
