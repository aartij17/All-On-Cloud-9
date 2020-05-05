package replica

import (
	"All-On-Cloud-9/common"
	"encoding/json"
	"fmt"
	"github.com/nats-io/nats.go"
	"All-On-Cloud-9/messenger"
	log "github.com/Sirupsen/logrus"
	"context"
	"os"
	"os/signal"
)

type Replica struct {
	Graph       string // STUB. For now, the DAG will just be a string
	NumReplicas int
}

func (replica *Replica) HandleReceive(message *common.MessageEvent) string {
	replica.AddDepsToGraph(message)
	return replica.ExecVertices()
}

// add the dependency to the graph
func (replica *Replica) AddDepsToGraph(message *common.MessageEvent) {
	fmt.Println("add dependency to graph")
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

func StartReplica(ctx context.Context) {
	nc, err := messenger.NatsConnect(ctx)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Error("error Replica connecting to nats server")
		return
	}

	rep := Replica{}
	go func(nc *nats.Conn, rep *Replica) {
		_, err = nc.Subscribe(common.ProposerToReplica, func(m *nats.Msg) {
			fmt.Println("Received proposer to replica")
			data := common.MessageEvent{}
			json.Unmarshal(m.Data, &data)
			newMessage := rep.HandleReceive(&data)
			sentMessage, err := json.Marshal(&newMessage)
			// Respond back to the client
			if err == nil {
				fmt.Println("leader can publish a message to deps")
				err = nc.Publish(common.NATS_CONSENSUS_DONE, sentMessage)
				if err != nil {
					log.WithFields(log.Fields{
						"error": err.Error(),
					}).Error("error publish NATS_CONSENSUS_DONE")
				}
			} else {
				fmt.Println("json marshal failed")
				fmt.Println(err.Error())
			}

		})
		if err != nil {
			log.WithFields(log.Fields{
				"error": err.Error(),
			}).Error("error subscribe ProposerToReplica")
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
