package node

import (
	"All-On-Cloud-9/common"
	"fmt"
)

var (
	id_count = 0
)

type Leader struct {
	Index int
}

// func Union(a, b []*Vertex, m map[*Vertex]bool) []*Vertex {
func Union(messages []*common.MessageEvent) common.MessageEvent {
	
	newMessage := common.MessageEvent{}

	if len(messages) > 0 {
		var m map[*common.Vertex]bool
		deps := messages[0].Deps

		for _, item := range deps {
				m[item] = true
		}

		for _, mes := range messages[1:] {
			for _, item := range mes.Deps {
				if _, ok := m[item]; !ok {
						deps = append(deps, item)
						m[item] = true
				}
			}	
		}

		newMessage.VertexId = messages[0].VertexId
		newMessage.Message = messages[0].Message
		newMessage.Deps = deps
	}

	return newMessage
}

func (leader *Leader) HandleReceiveCommand(message string) {
	v := common.Vertex{leader.Index, id_count}
	id_count += 1
	newMessageEvent := common.MessageEvent{&v, message, []*common.Vertex{}}
	fmt.Println(newMessageEvent.Message)

	// send message to dependency
}

func (leader *Leader) HandleReceiveDeps(messages []*common.MessageEvent) {
	// deps is a list of vertices
	if len(messages) > 0 {
		newMessageEvent := Union(messages)
		fmt.Println(newMessageEvent.Message)
	}
	// send message to proposer
}