package main

import (
	"All-On-Cloud-9/common"
	"All-On-Cloud-9/config"
	"All-On-Cloud-9/messenger"
	"All-On-Cloud-9/pbft"
	"context"
)

func testPbft(ctx context.Context, id int) {
	nc, _ := messenger.NatsConnect(ctx)
	node := pbft.NewPbftNode(ctx, nc, "TEST", 1, 4, id)
	
	if id == 0 {
		node.MessageIn <- &common.Transaction{
			LocalXNum:  "A",
			GlobalXNum: "d",
			Type:       "LOCAL",
			TxnId:      "",
			ToId:       "",
			FromId:     "",
			CryptoHash: "",
			TxnType:    "",
			Clock:      nil,
		}
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
