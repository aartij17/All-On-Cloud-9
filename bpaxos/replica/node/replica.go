package replica

import (
	"All-On-Cloud-9/common"
	"All-On-Cloud-9/messenger"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"

	log "github.com/Sirupsen/logrus"
	"github.com/nats-io/nats.go"
	// "github.com/hashicorp/terraform/dag"
)

// var (
// 	m = make(map[string]string)
// )

type Replica struct {
	NumReplicas int
}

// type GraphNode struct {
// 	VertexId int
// 	LeaderIndex int
// 	Message string
// }

func (replica *Replica) HandleReceive(message *common.MessageEvent) string {
	replica.AddDepsToGraph(message)
	return replica.ExecVertices()
}

// add the dependency to the graph
func (replica *Replica) AddDepsToGraph(message *common.MessageEvent) {
	fmt.Println("add dependency to graph")
	// newNode := &GraphNode {VertexId:message.VertexId.Id, LeaderIndex:message.VertexId.Index, Message:message.Message}
	// strkey := fmt.Sprintf("%d:%d", message.VertexId.Id, message.VertexId.Index)
	// m[strkey] = message.Message
	// replica.Graph.add(strkey)
	// for index, dep := range message.Deps {
	// 	strkeyDep := fmt.Sprintf("%d:%d", dep.VertexId.Id, dep.VertexId.Index)
	// 	replica.Graph.Connect(BasicEdge(strkeyDep, strkey))
	// }

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

func ProcessReplicaMessage(m *nats.Msg, nc *nats.Conn, ctx context.Context, rep *Replica) {
	fmt.Println("Received proposer to replica")
	data := common.MessageEvent{}
	json.Unmarshal(m.Data, &data)
	// newMessage := rep.HandleReceive(&data)
	// sentMessage, err := json.Marshal(&newMessage)
	// // Respond back to the client
	// if err == nil {
	// 	fmt.Println("leader can publish a message to deps")
	// 	messenger.PublishNatsMessage(ctx, nc, common.NATS_CONSENSUS_DONE_MSG, sentMessage)

	// } else {
	// 	fmt.Println("json marshal failed")
	// 	fmt.Println(err.Error())
	// }
}

func StartReplica(ctx context.Context, nc *nats.Conn) {
	rep := Replica{}
	go func(nc *nats.Conn, rep *Replica) {
		NatsMessage := make(chan *nats.Msg)
		err := messenger.SubscribeToInbox(ctx, nc, common.PROPOSER_TO_REPLICA, NatsMessage, false)

		if err != nil {
			log.WithFields(log.Fields{
				"error": err.Error(),
			}).Error("error subscribe PROPOSER_TO_REPLICA")
		}
		var (
			natsMsg *nats.Msg
		)
		for {
			select {
			case natsMsg = <-NatsMessage:
				ProcessReplicaMessage(natsMsg, nc, ctx, rep)
			}
		}
	}(nc, &rep)

	signalChan := make(chan os.Signal, 1)
	cleanupDone := make(chan bool)
	signal.Notify(signalChan, os.Interrupt)
	go func() {
		for _ = range signalChan {
			log.Info("Received an interrupt, stopping all connections...")
			//cancel()
			cleanupDone <- true
		}
	}()
	<-cleanupDone
}
