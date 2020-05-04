package leadernode

import (
	"All-On-Cloud-9/common"
	"encoding/json"
	"fmt"
	"github.com/nats-io/nats.go"
)

var (
	id_count = 0
)

type Leader struct {
	Index    int
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

func (leader *Leader) HandleReceiveCommand(message string) common.MessageEvent {
	v := common.Vertex{leader.Index, id_count}
	id_count += 1
	newMessageEvent := common.MessageEvent{&v, message, []*common.Vertex{}}
	// fmt.Println(newMessageEvent.Message)  // STUB
	return newMessageEvent
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

func (leader *Leader) GetMessagesLen() int {
	return len(leader.messages)
}

func StartLeader(leaderindex int) {
	l := NewLeader(leaderindex) // Hard Coded User Id.
	socket := common.Socket{}
	_ = socket.Connect(nats.DefaultURL)

	go func(leader *Leader, socket *common.Socket) {
		socket.Subscribe(common.DepsToLeader, func(m *nats.Msg) {
			fmt.Println("Received deps to leader")
			data := common.MessageEvent{}
			json.Unmarshal(m.Data, &data)
			l.AddToMessages(&data)
			if l.GetMessagesLen() > common.F {
				newMessageEvent := l.HandleReceiveDeps()

				sentMessage, err := json.Marshal(&newMessageEvent)
				if err == nil {
					fmt.Println("leader can publish a message to proposer")
					socket.Publish(common.LeaderToProposer, sentMessage)
				} else {
					fmt.Println("json marshal failed")
					fmt.Println(err.Error())
				}
				// should we flush when it fails?
				l.FlushMessages()
			}
		})
	}(&l, &socket)

	go func(leader *Leader, socket *common.Socket) {
		socket.Subscribe(common.NATS_CONSENSUS_INITIATE_MSG, func(m *nats.Msg) {
			fmt.Println("Received client to leader")
			newMessage := leader.HandleReceiveCommand(string(m.Data))
			sentMessage, err := json.Marshal(&newMessage)
			if err == nil {
				fmt.Println("leader can publish a message to deps")
				socket.Publish(common.LeaderToDeps, sentMessage)
			} else {
				fmt.Println("json marshal failed")
				fmt.Println(err.Error())
			}
		})
	}(&l, &socket)
	common.HandleInterrupt()
}
