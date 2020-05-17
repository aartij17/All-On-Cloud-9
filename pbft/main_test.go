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
	config.LoadConfig(ctx, "./config/config.json")

	for i := 0; i < 4; i++ {
		go func() {
			nc, _ := messenger.NatsConnect(ctx)
			node := NewPbftNode(ctx, nc, "TEST", 1, 4, 0, 1, i, 0)

			dummyTxn := &common.Transaction{
				Type: LOCAL,
			}
			if i == 0 {
				node.MessageIn <- dummyTxn
			}
			txn := <- node.MessageOut
			if !reflect.DeepEqual(txn, dummyTxn) {
				t.Error("Wrong LOCAL outcome, transaction is:")
				t.Error(txn)
			}
			done <- true
		}()
	}

	select {
	case <-timeout:
		t.Error("Local caper timed out")
	case <-done:
	}
}

func TestGlobal(t *testing.T) {
	timeout := time.After(3 * TIMEOUT * time.Second)
	done := make(chan bool)

	ctx, _ := context.WithCancel(context.Background())
	config.LoadConfig(ctx, "./config/config.json")

	for i := 0; i < 4; i++ {
		go func() {
			nc, _ := messenger.NatsConnect(ctx)
			node := NewPbftNode(ctx, nc, "TEST" + strconv.Itoa(i), 0, 1, 1, 4, 0, i)

			dummyTxn := &common.Transaction{
				Type: GLOBAL,
			}
			if i == 0 {
				node.MessageIn <- dummyTxn
			}
			txn := <- node.MessageOut
			if !reflect.DeepEqual(txn, dummyTxn) {
				t.Error("Wrong GLOBAL outcome, transaction is:")
				t.Error(txn)
			}
			done <- true
		}()
	}

	select {
	case <-timeout:
		t.Error("Global caper timed out")
	case <-done:
	}
}