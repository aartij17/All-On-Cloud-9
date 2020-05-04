package depsnode

import (
	"All-On-Cloud-9/common"
	"encoding/json"
	"fmt"
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
	key := Messagekey{message.VertexId, message.Message}
	if !depsServiceNode.CmdsMap[key] {
		depsServiceNode.Cmds = append(depsServiceNode.Cmds, message)
		depsServiceNode.CmdsMap[key] = true
	}

	// Now send a new message with the calculated dependencies back to the leader
	newMessage := common.MessageEvent{message.VertexId, message.Message, deps}
	fmt.Println("Send new message to leader: %s", newMessage.Message) // STUB: stop the compiler from complaining until implement the real send
	return newMessage

}

func (depsServiceNode *DepsServiceNode) Stub() {
	fmt.Println("Dependency Service Node: STUB PLS REMOVE")
}

func StartDependencyService() {
	dependency_node := NewDepsServiceNode()
	// dependency_node.Stub()
	go func(dep_node *DepsServiceNode) {
		socket := common.Socket{}
		_ = socket.Connect(nats.DefaultURL)
		socket.Subscribe(common.LeaderToDeps, func(m *nats.Msg) {
			fmt.Println("Received leader to deps")
			data := common.MessageEvent{}
			json.Unmarshal(m.Data, &data)
			newMessage := dep_node.HandleReceive(&data)
			sentMessage, err := json.Marshal(&newMessage)

			if err == nil {
				fmt.Println("deps can publish a message to leader")
				socket.Publish(common.DepsToLeader, sentMessage)
			} else {
				fmt.Println("json marshal failed")
				fmt.Println(err.Error())
			}

		})
	}(&dependency_node)
	common.HandleInterrupt()
}
