package main

import (
	"All-On-Cloud-9/bpaxos"
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"

	"github.com/nats-io/nats.go"
	log "github.com/sirupsen/logrus"
)

func main() {
	var (
		nodeType    int
		nodeId      string
		err         error
		numberProps string
	)
	flag.IntVar(&nodeType, "nodetype", 0, "node type(0 - leader, 1 - proposer, 2 - consensus, 3 - replica)")
	flag.StringVar(&numberProps, "numberProps", "1", "number of proposers")
	flag.StringVar(&nodeId, "nodeId", "0", "node ID(0 - leader, 1 - proposer, 2, 3, 4 - consensus)")
	fmt.Println(os.Args)

	flag.Parse()

	ctx, _ := context.WithCancel(context.Background())
	natsOptions := nats.Options{
		Servers:        []string{"nats://0.0.0.0:4222"},
		AllowReconnect: true,
	}
	nc, err := natsOptions.Connect()
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Error("error connecting to nats server")
		return
	}
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Error("error connecting to nats")
		return
	}

	switch nodeType {
	// leader
	case 0:
		os.Setenv("NUM_PROP", numberProps)
		fmt.Println("NUM_PROP: " + numberProps)
		bpaxos.SetupBPaxos(ctx, nc, false, true, false, false)
	// proposer
	case 1:
		os.Setenv("PROP_ID", nodeId)
		fmt.Println("PROP_ID: " + nodeId)
		bpaxos.SetupBPaxos(ctx, nc, false, false, true, false)
	// consensus
	case 2:
		bpaxos.SetupBPaxos(ctx, nc, true, false, false, false)
	// replica
	case 3:
		bpaxos.SetupBPaxos(ctx, nc, false, false, false, true)
	}

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
