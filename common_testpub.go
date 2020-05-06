package main

import (
	"fmt"
	"All-On-Cloud-9/common"
	"github.com/nats-io/nats.go"
	// "encoding/json"
	log "github.com/Sirupsen/logrus"
)

type Socket struct {
	conn *nats.Conn
}

func (s *Socket) Connect(address string) bool {
	var err error

	natsOptions := nats.Options{
		Servers:        []string{address},
		AllowReconnect: true,
		MaxReconnect:   3,
		ReconnectWait:  10,
		//// TODO: these callbacks can be set in order to allow nats to perform failure handling.
		//ClosedCB:          nil,
		//DisconnectedErrCB: nil,
	}

	s.conn, err = natsOptions.Connect()
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Error("error connecting to the NATS server")
		return false
	}
	return true
}

func (s *Socket) Subscribe(subject string, response nats.MsgHandler) {
	_, err := s.conn.Subscribe(subject, response)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Error("error subscribing to the subject")
	}
}

func (s *Socket) Publish(subject string, request []byte) {
	err := s.conn.Publish(subject, request)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Error("error publishing the nats request")
	}
}


func main() {
	// STUB FILE REMOVE LATER
	// v0 := common.Vertex{0,0}
	// message1 := common.MessageEvent{&v0, "Hello", []*common.Vertex{&v0, &v0}}
	// sentMessage, err := json.Marshal(&message1)
	// if err == nil {
	// 	fmt.Println("no marshal booboo")
	// } else {
	// 	fmt.Println(err.Error())
	// }
	socket := common.Socket{}
	_ = socket.Connect(nats.DefaultURL)
	socket.Publish(common.NATS_CONSENSUS_INITIATE_MSG, []byte("stub"))
	// socket.Publish(common.LeaderToProposer, sentMessage)
	// socket.Publish(common.ProposerToReplica, sentMessage)

	socket.Subscribe(common.LeaderToDeps, func(m *nats.Msg) {
		fmt.Println("Received")
	})
	fmt.Println("YAY")
}