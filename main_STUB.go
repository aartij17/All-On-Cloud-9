package main

import (
	"All-On-Cloud-9/bpaxos"
	"All-On-Cloud-9/common"
	"All-On-Cloud-9/messenger"
	"github.com/nats-io/nats.go"
	"context"
	"All-On-Cloud-9/config"
	"time"
	"fmt"
	"os/signal"
	"os"
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
	ctx, _ := context.WithCancel(context.Background())

	config.LoadConfig(ctx, "config/config.json")

	nc, err := messenger.NatsConnect(ctx)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Error("error Replica connecting to nats server")
		return
	}

	bpaxos.SetupBPaxos(ctx, nc, true, true, true, true, true)
	time.Sleep(20000 * time.Millisecond)
	fmt.Println("Starting. Please work")
	socket := Socket{}
	_ = socket.Connect(nats.DefaultURL)
	socket.Publish(common.NATS_CONSENSUS_INITIATE_MSG, []byte("bitch"))
	// socket.Publish(LEADER_TO_PROPOSER, sentMessage)
	// socket.Publish(PROPOSER_TO_REPLICA, sentMessage)
	socket.Subscribe(common.NATS_CONSENSUS_DONE, func(m *nats.Msg) {
		fmt.Println("Received")
	})
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