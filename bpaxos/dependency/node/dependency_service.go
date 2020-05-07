package depsnode

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
)

// Private datatype
type Messagekey struct {
	// This will be used for the key into the map
	VertexId *common.Vertex
	Message  string
}

type DepsServiceNode struct {
	Cmds    []*common.MessageEvent // Array of commands
	CmdsMap map[Messagekey]bool    // Keep a map for quick lookup of which message are already part of Cmds
}

func NewDepsServiceNode() DepsServiceNode {
	depsServiceNode := DepsServiceNode{}
	depsServiceNode.CmdsMap = make(map[Messagekey]bool)
	depsServiceNode.Cmds = nil
	return depsServiceNode
}

func (depsServiceNode *DepsServiceNode) ComputeConflictingMessages(message *common.MessageEvent) []*common.Vertex {
	// Find and all commands that conflict with message and add them to deps
	deps := []*common.Vertex{}

	return deps
}

func (depsServiceNode *DepsServiceNode) HandleReceive(message *common.MessageEvent) common.MessageEvent {
	deps := depsServiceNode.ComputeConflictingMessages(message)

	// Append message to Cmds if it is not already inside
	key := Messagekey{message.VertexId, string(message.Message)}
	if !depsServiceNode.CmdsMap[key] {
		depsServiceNode.Cmds = append(depsServiceNode.Cmds, message)
		depsServiceNode.CmdsMap[key] = true
	}

	// Now send a new message with the calculated dependencies back to the leader
	newMessage := common.MessageEvent{message.VertexId, message.Message, deps}
	fmt.Println("Send new message to leader: %s", newMessage.Message) // STUB: stop the compiler from complaining until implement the real send
	return newMessage

}

func ProcessDepMessage(m *nats.Msg, nc *nats.Conn, ctx context.Context, dep_node *DepsServiceNode) {
	fmt.Println("Received leader to deps")
	data := common.MessageEvent{}
	json.Unmarshal(m.Data, &data)
	newMessage := dep_node.HandleReceive(&data)
	sentMessage, err := json.Marshal(&newMessage)

	if err == nil {
		fmt.Println("deps can publish a message to leader")
		messenger.PublishNatsMessage(ctx, nc, common.DEPS_TO_LEADER, sentMessage)

	} else {
		fmt.Println("json marshal failed")
		fmt.Println(err.Error())
	}
}

func StartDependencyService(ctx context.Context, nc *nats.Conn) {
	dependency_node := NewDepsServiceNode()
	// dependency_node.Stub()
	go func(nc *nats.Conn, dep_node *DepsServiceNode) {
		NatsMessage := make(chan *nats.Msg)
		err := messenger.SubscribeToInbox(ctx, nc, common.LEADER_TO_DEPS, NatsMessage)

		if err != nil {
			log.WithFields(log.Fields{
				"error": err.Error(),
			}).Error("error subscribe LEADER_TO_DEPS")
		}

		var (
			natsMsg *nats.Msg
		)
		for {
			select {
			case natsMsg = <-NatsMessage:
				ProcessDepMessage(natsMsg, nc, ctx, dep_node)
			}
		}
	}(nc, &dependency_node)

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
