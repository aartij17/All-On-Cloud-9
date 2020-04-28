package leadernode

import (
	"All-On-Cloud-9/common"
	"fmt"
)

var (
	id_count = 0
)

type Leader struct {
	Index int
	messages []*common.MessageEvent
}

func NewLeader(index int) Leader {
	l := Leader{}
	l.Index = index
	l.FlushMessages()
	return l
}

// func Union(a, b []*Vertex, m map[*Vertex]bool) []*Vertex {
func (leader *Leader) Union() common.MessageEvent {
	
	newMessage := common.MessageEvent{}

	if len(leader.messages) > 0 {

		m := map[*common.Vertex]bool{}

		deps := leader.messages[0].Deps

		for _, item := range deps {
			m[item] = true
		}

		for _, mes := range leader.messages[1:] {
			for _, item := range mes.Deps {
				if _, ok := m[item]; !ok {
						deps = append(deps, item)
						m[item] = true
				}
			}	
		}

		newMessage.VertexId = leader.messages[0].VertexId
		newMessage.Message = leader.messages[0].Message
		newMessage.Deps = deps
	}

	return newMessage
}

func (leader *Leader) HandleReceiveCommand(message string) {
	v := common.Vertex{leader.Index, id_count}
	id_count += 1
	newMessageEvent := common.MessageEvent{&v, message, []*common.Vertex{}}
	fmt.Println(newMessageEvent.Message)  // STUB

	// send message to dependency
}

func (leader *Leader) HandleReceiveDeps() common.MessageEvent {
	
	newMessageEvent := leader.Union()
	return newMessageEvent
	
}

func (leader *Leader) AddToMessages(message *common.MessageEvent) {
	leader.messages = append(leader.messages, message)
}

func (leader *Leader) FlushMessages() {
	leader.messages = nil
}

func (leader *Leader) GetMessagesLen() int{
	return len(leader.messages)
}