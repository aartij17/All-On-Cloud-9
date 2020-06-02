package main

import (
	"All-On-Cloud-9/common"
	"github.com/nats-io/nats.go"
	"time"
	"fmt"
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
	cleanupDone := make(chan bool)
	numConsensusDone := 0

	fmt.Println("Starting. Please work")
	socket := Socket{}
	_ = socket.Connect(nats.DefaultURL)

	numMessages := 50
	
	socket.Subscribe(common.NATS_CONSENSUS_DONE_MSG, func(m *nats.Msg) {
		numConsensusDone += 1
		fmt.Println("Received")
		if numConsensusDone == numMessages {
			cleanupDone <- true
		}
	})
	
	start := time.Now()
	for i := 0; i < numMessages; i++ {
		go socket.Publish(common.NATS_CONSENSUS_INITIATE_MSG, []byte("bitch"))
	}

	<-cleanupDone
	elapsed := time.Since(start)
	fmt.Printf("FINISHED! %s", elapsed)
}
