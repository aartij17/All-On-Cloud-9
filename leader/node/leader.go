package node

import (
	"All-On-Cloud-9/common"
)

var (
	id_count = 0
)

type Leader struct {
	Index int
}

func Union(a, b []*Vertex, m map[*Vertex]bool) []*Vertex {

	for _, item := range a {
			m[item] = true
	}

	for _, item := range b {
			if _, ok := m[item]; !ok {
					a = append(a, item)
			}
	}
	return a
}

func Union(a, messages[]*MessageEvent ) []*Vertex {
	m := map[*Vertex]bool
	
	for _, item := range a {
		m[item] = true
	}

	for _, mes := range messages {
		for _, item := range mes.Deps {
			if _, ok := m[item]; !ok {
					a = append(a, item)
					m[item] = true
			}
		}	
	}

	return a
}

func (leader *Leader) handleReceiveCommand(message string) {
	v := Vertex{leader.Index, id_count}
	id_count += 1
	newMessageEvent := MessageEvent{v, message, []*Vertex{}}
	// send message to dependency
}

func (leader *Leader) handleReceiveDeps(messages []*MessageEvent) {
	// deps is a list of vertices
	deps := Union(messages[0].Deps, messages[1:])
	newMessageEvent := MessageEvent{messages[0].VertexId, messages[0].Message, deps}
	// send message to proposer
}