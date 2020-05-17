package main

import (
	"All-On-Cloud-9/common"
	"All-On-Cloud-9/config"
	"All-On-Cloud-9/messenger"
	"All-On-Cloud-9/pbft"
	"context"
	"github.com/Sirupsen/logrus"
	"strconv"
)

func testPbft(ctx context.Context, id int) {
	nc, _ := messenger.NatsConnect(ctx)
	node := pbft.NewPbftNode(ctx, nc, "TEST" + strconv.Itoa(id), 0, 1, 1, 4, 0, id)

	if id == 0 {
		node.MessageIn <- &common.Transaction{
			LocalXNum:  "A",
			GlobalXNum: "d",
			Type:       "GLOBAL",
			TxnId:      "",
			ToId:       "",
			FromId:     "",
			CryptoHash: "",
			TxnType:    "",
			Clock:      nil,
		}
	}
	for {
		txn := <- node.MessageOut
		logrus.WithField("txn", *txn).Info("Caper Out")
	}
}

func main() {
	ctx, _ := context.WithCancel(context.Background())
	config.LoadConfig(ctx, "./config/config.json")

	for i := 0; i < 4; i++ {
		go testPbft(ctx, i)
	}

	sleep := make(chan bool)
	<- sleep
}
