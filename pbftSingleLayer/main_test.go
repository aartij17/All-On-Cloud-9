package pbftSingleLayer

import (
	"All-On-Cloud-9/common"
	"All-On-Cloud-9/config"
	"All-On-Cloud-9/messenger"
	"reflect"
	"testing"
	"time"

	"context"
	"strconv"
)

func setupNode(ctx context.Context, t *testing.T, done chan bool) func(int, int) {
	return func(id int, appId int) {
		nc, _ := messenger.NatsConnect(ctx)
		node := NewPbftNode(ctx, nc, "APP_"+strconv.Itoa(appId), id, appId)

		dummyTxn := common.Transaction{
			TxnType: GLOBAL,
		}
		if id == 0 && appId == 0 {
			node.MessageIn <- dummyTxn
		}
		txn := <-node.MessageOut
		if !reflect.DeepEqual(txn, dummyTxn) {
			t.Error("Wrong GLOBAL outcome, transaction is:")
			t.Error(txn)
			t.Error("Expected is:")
			t.Error(dummyTxn)
		}
		done <- true
	}
}

func TestGlobalSingleNodeApp(t *testing.T) {
	timeout := time.After(3 * TIMEOUT * time.Second)
	done := make(chan bool)

	ctx, _ := context.WithCancel(context.Background())
	config.LoadConfig(ctx, "../config/config.json")

	for i := 0; i < 4; i++ {
		go setupNode(ctx, t, done)(0, i)
	}

	for i := 0; i < 4; i++ {
		select {
		case <-timeout:
			t.Error("Global pbft timed out")
		case <-done:
		}
	}
}

func TestGlobalOneMultipleNodeApp(t *testing.T) {
	timeout := time.After(3 * TIMEOUT * time.Second)
	//timeout := time.After(500 * time.Second)
	done := make(chan bool)
	ctx, _ := context.WithCancel(context.Background())
	config.LoadConfig(ctx, "../config/config.json")

	const AppCount = 4
	const NodePerFirstApp = 4
	for j := 0; j < NodePerFirstApp; j++ {
		go setupNode(ctx, t, done)(j, 0)
	}
	for i := 1; i < AppCount; i++ {
		go setupNode(ctx, t, done)(0, i)
	}

	for i := 0; i < AppCount+NodePerFirstApp-1; i++ {
		select {
		case <-timeout:
			t.Error("Global pbft timed out")
		case <-done:
		}
	}
	time.Sleep(500 * time.Millisecond)
}

func TestGlobalMultipleNodeApp(t *testing.T) {
	//timeout := time.After(3 * TIMEOUT * time.Second)
	timeout := time.After(500 * time.Second)
	done := make(chan bool)

	ctx, _ := context.WithCancel(context.Background())
	config.LoadConfig(ctx, "../config/config.json")

	const AppCount = 4
	const NodePerApp = 4
	for i := 0; i < AppCount; i++ {
		for j := 0; j < NodePerApp; j++ {
			go setupNode(ctx, t, done)(j, i)
		}
	}

	for i := 0; i < AppCount*NodePerApp; i++ {
		select {
		case <-timeout:
			t.Error("Global pbft timed out")
		case <-done:
		}
	}
	time.Sleep(500 * time.Millisecond)
}
