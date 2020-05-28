package pbft

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

func TestLocal(t *testing.T) {
	timeout := time.After(3 * TIMEOUT * time.Second)
	done := make(chan bool)

	ctx, _ := context.WithCancel(context.Background())
	config.LoadConfig(ctx, "../config/config.json")

	for i := 0; i < 4; i++ {
		go func(id int) {
			nc, _ := messenger.NatsConnect(ctx)
			node := NewPbftNode(ctx, nc, "APP", 1, 4, 0, 1, id, 0)

			dummyTxn := common.Transaction{
				TxnType: LOCAL,
			}
			if id == 0 {
				node.MessageIn <- dummyTxn
			}
			txn := <-node.MessageOut
			if !reflect.DeepEqual(txn, dummyTxn) {
				t.Error("Wrong LOCAL outcome, transaction is:")
				t.Error(txn)
				t.Error("Expected is:")
				t.Error(dummyTxn)
			}
			done <- true
		}(i)
	}

	select {
	case <-timeout:
		t.Error("Local pbft timed out")
	case <-done:
	}
}

func TestGlobalSingleNodeApp(t *testing.T) {
	timeout := time.After(3 * TIMEOUT * time.Second)
	done := make(chan bool)

	ctx, _ := context.WithCancel(context.Background())
	config.LoadConfig(ctx, "../config/config.json")

	for i := 0; i < 4; i++ {
		go func(id int) {
			nc, _ := messenger.NatsConnect(ctx)
			node := NewPbftNode(ctx, nc, "APP_"+strconv.Itoa(id), 0, 1, 1, 4, 0, id)
			go PipeInHierarchicalLocalConsensus(node)

			dummyTxn := common.Transaction{
				TxnType: GLOBAL,
			}
			if id == 0 {
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
		}(i)
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
	//timeout := time.After(3 * TIMEOUT * time.Second)
	timeout := time.After(500 * time.Second)
	done := make(chan bool)

	ctx, _ := context.WithCancel(context.Background())
	config.LoadConfig(ctx, "../config/config.json")

	const AppCount = 4
	const NodePerApp = 4
	for j := 0; j < NodePerApp; j++ {
		go func(id int, appId int) {
			nc, _ := messenger.NatsConnect(ctx)
			node := NewPbftNode(ctx, nc, "APP_0", 1, 4, 1, 4, id, appId)
			go PipeInHierarchicalLocalConsensus(node)

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
		}(j, 0)
	}
	for i := 1; i < AppCount; i++ {
		go func(id int, appId int) {
			nc, _ := messenger.NatsConnect(ctx)
			node := NewPbftNode(ctx, nc, "APP_"+strconv.Itoa(appId), 0, 1, 1, 4, id, appId)
			go PipeInHierarchicalLocalConsensus(node)

			dummyTxn := common.Transaction{
				TxnType: GLOBAL,
			}
			txn := <-node.MessageOut
			if !reflect.DeepEqual(txn, dummyTxn) {
				t.Error("Wrong GLOBAL outcome, transaction is:")
				t.Error(txn)
				t.Error("Expected is:")
				t.Error(dummyTxn)
			}
			done <- true
		}(0, i)
	}

	for i := 0; i < AppCount+NodePerApp-1; i++ {
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
			go func(id int, appId int) {
				nc, _ := messenger.NatsConnect(ctx)
				node := NewPbftNode(ctx, nc, "APP_"+strconv.Itoa(appId), 1, 4, 1, 4, id, appId)
				go PipeInHierarchicalLocalConsensus(node)

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
			}(j, i)
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
