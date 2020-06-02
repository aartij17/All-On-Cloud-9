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

func TestTHLocal(t *testing.T) {
	timeout := time.After(3 * TIMEOUT * time.Second)
	done := make(chan bool)

	ctx, _ := context.WithCancel(context.Background())
	config.LoadConfig(ctx, "../config/config.json")

	for i := 0; i < 3; i++ {
		go func(id int) {
			nc, _ := messenger.NatsConnect(ctx)
			node := NewPbftNode(ctx, nc, "APP", 1, 3, 0, 1, id, 0, true)

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

func TestTHGlobalSingleNodeApp(t *testing.T) {
	timeout := time.After(3 * TIMEOUT * time.Second)
	done := make(chan bool)

	ctx, _ := context.WithCancel(context.Background())
	config.LoadConfig(ctx, "../config/config.json")

	for i := 0; i < 3; i++ {
		go func(id int) {
			nc, _ := messenger.NatsConnect(ctx)
			node := NewPbftNode(ctx, nc, "APP_"+strconv.Itoa(id), 0, 1, 1, 3, 0, id, true)
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

	for i := 0; i < 3; i++ {
		select {
		case <-timeout:
			t.Error("Global pbft timed out")
		case <-done:
		}
	}
}

func TestTHGlobalOneMultipleNodeApp(t *testing.T) {
	timeout := time.After(3 * TIMEOUT * time.Second)
	//timeout := time.After(500 * time.Second)
	done := make(chan bool)

	ctx, _ := context.WithCancel(context.Background())
	config.LoadConfig(ctx, "../config/config.json")

	const AppCount = 3
	const NodePerApp = 3
	for j := 0; j < NodePerApp; j++ {
		go func(id int, appId int) {
			nc, _ := messenger.NatsConnect(ctx)
			node := NewPbftNode(ctx, nc, "APP_0", 1, NodePerApp, 1, AppCount, id, appId, true)
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
			node := NewPbftNode(ctx, nc, "APP_"+strconv.Itoa(appId), 0, 1, 1, AppCount, id, appId, true)
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
}

func TestTHGlobalAndLocalOneMultipleNodeApp(t *testing.T) {
	timeout := time.After(3 * TIMEOUT * time.Second)
	//timeout := time.After(500 * time.Second)
	done := make(chan bool)

	ctx, _ := context.WithCancel(context.Background())
	config.LoadConfig(ctx, "../config/config.json")

	const AppCount = 3
	const NodePerFirstApp = 3
	dummyTxn := common.Transaction{
		TxnType:   GLOBAL,
		Timestamp: 1,
	}
	dummyTxn2 := common.Transaction{
		TxnType:   GLOBAL,
		Timestamp: 3,
	}
	dummyLocalTxn := common.Transaction{
		TxnType:   LOCAL,
		Timestamp: 2,
	}
	dummyLocalTxn2 := common.Transaction{
		TxnType:   LOCAL,
		Timestamp: 4,
	}
	Txns := [4]common.Transaction{
		dummyTxn,
		dummyTxn2,
		dummyLocalTxn,
		dummyLocalTxn2,
	}
	for j := 0; j < NodePerFirstApp; j++ {
		go func(id int, appId int) {
			nc, _ := messenger.NatsConnect(ctx)
			node := NewPbftNode(ctx, nc, "APP_0", 1, NodePerFirstApp, 1, AppCount, id, appId, true)
			go PipeInHierarchicalLocalConsensus(node)

			for i, _txn := range Txns {
				if id == 0 && appId == 0 {
					println(i)
					//println(_txn)
					node.MessageIn <- _txn
				}
				txn := <-node.MessageOut
				if !reflect.DeepEqual(txn, _txn) {
					t.Error("Wrong GLOBAL outcome, transaction is:")
					t.Error(txn)
					t.Error("Expected is:")
					t.Error(_txn)
				}
				done <- true
			}
		}(j, 0)
	}
	for i := 1; i < AppCount; i++ {
		go func(id int, appId int) {
			nc, _ := messenger.NatsConnect(ctx)
			node := NewPbftNode(ctx, nc, "APP_"+strconv.Itoa(appId), 0, 1, 1, AppCount, id, appId, true)
			go PipeInHierarchicalLocalConsensus(node)

			for _, _txn := range Txns {
				if _txn.TxnType == GLOBAL {
					txn := <-node.MessageOut
					if !reflect.DeepEqual(txn, _txn) {
						t.Error("Wrong GLOBAL outcome, transaction is:")
						t.Error(txn)
						t.Error("Expected is:")
						t.Error(_txn)
					}
					done <- true
				}
			}
		}(0, i)
	}

	const (
		GlobalCnt = 2
		LocalCnt  = 2
	)
	for i := 0; i < GlobalCnt*AppCount+(GlobalCnt+LocalCnt)*NodePerFirstApp-GlobalCnt; i++ {
		select {
		case <-timeout:
			t.Error("Global pbft timed out")
		case <-done:
		}
	}
}

func TestTHGlobalMultipleNodeApp(t *testing.T) {
	timeout := time.After(3 * TIMEOUT * time.Second)
	//timeout := time.After(500 * time.Second)
	done := make(chan bool)

	ctx, _ := context.WithCancel(context.Background())
	config.LoadConfig(ctx, "../config/config.json")

	const AppCount = 3
	const NodePerApp = 3
	for i := 0; i < AppCount; i++ {
		for j := 0; j < NodePerApp; j++ {
			go func(id int, appId int) {
				nc, _ := messenger.NatsConnect(ctx)
				node := NewPbftNode(ctx, nc, "APP_"+strconv.Itoa(appId), 1, NodePerApp, 1, AppCount, id, appId, true)
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
}
