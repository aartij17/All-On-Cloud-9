package main

import (
	"testing"
	"encoding/json"
	"All-On-Cloud-9/common"
	"All-On-Cloud-9/server/application"
	"All-On-Cloud-9/server/blockchain"
	"github.com/nats-io/nats.go"
	"fmt"
)

func TestCarrierContractSuccess(t *testing.T) {
	carrier := application.Carrier{
		ContractValid: make(chan bool),
		MsgChannel:    make(chan *nats.Msg),
	}

	carrierReq := application.CarrierClientRequest{
		ToApp: "whatever",
		Fee: 7,
	}

	jTxn, _ := json.Marshal(&carrierReq)

	txn := &common.Transaction{
		TxnBody: jTxn,
		FromApp: "CARRIER",
		ToApp:   carrierReq.ToApp,
		ToId:    "",
		FromId:  "",
		TxnType: "",
		Clock:   nil,
	}

	block := blockchain.Block {
		IsGenesis: false,
		BlockId: "",
		ViewType: "",
		CryptoHash: "",
		Transaction: txn,
		InitiatorNode: "",
		Clock: nil,
	}

	fmt.Println("start run carreir contract")

	carrier.RunCarrierContract(block)
	isValid := <-carrier.ContractValid
	if isValid == false {
		t.Errorf("run carrier contract failed")
	}
	fmt.Println("run carreir contract finished")
}