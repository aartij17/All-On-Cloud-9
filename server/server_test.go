package main

import (
	"All-On-Cloud-9/common"
	"All-On-Cloud-9/server/application"
	"All-On-Cloud-9/server/blockchain"
	"encoding/json"
	"testing"

	"github.com/nats-io/nats.go"
)

func TestCarrierContractSuccess(t *testing.T) {
	carrier := application.Carrier{
		ContractValid: make(chan bool),
		MsgChannel:    make(chan *nats.Msg),
	}

	carrierReq := application.CarrierClientRequest{
		ToApp: "whatever",
		Fee:   7,
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

	block := blockchain.Block{
		IsGenesis:     false,
		BlockId:       "",
		ViewType:      "",
		CryptoHash:    "",
		Transaction:   txn,
		InitiatorNode: "",
		Clock:         nil,
	}

	go carrier.RunCarrierContract(block)
	isValid := <-carrier.ContractValid
	if isValid == false {
		t.Errorf("run carrier contract failed")
	}
}

func TestCarrierContractFail(t *testing.T) {
	carrier := application.Carrier{
		ContractValid: make(chan bool),
		MsgChannel:    make(chan *nats.Msg),
	}

	carrierReq := application.CarrierClientRequest{
		ToApp: "whatever",
		Fee:   -7,
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

	block := blockchain.Block{
		IsGenesis:     false,
		BlockId:       "",
		ViewType:      "",
		CryptoHash:    "",
		Transaction:   txn,
		InitiatorNode: "",
		Clock:         nil,
	}

	go carrier.RunCarrierContract(block)
	isValid := <-carrier.ContractValid
	if isValid == true {
		t.Errorf("run carrier contract should have failed")
	}
}

func TestSupplierContractSuccess(t *testing.T) {
	supplier := application.Supplier{
		ContractValid: make(chan bool),
		MsgChannel:    make(chan *nats.Msg),
	}

	numUnitsToBuy := 5
	shippingService := "UPS"
	supplierReq := application.SupplierRequest{
		ToApp:           "",
		NumUnitsToBuy:   numUnitsToBuy,
		AmountPaid:      numUnitsToBuy*application.ManufacturerCostPerUnit + application.Rates[shippingService],
		ShippingService: shippingService,
		ShippingCost:    application.Rates[shippingService],
	}

	jTxn, _ := json.Marshal(&supplierReq)

	txn := &common.Transaction{
		TxnBody: jTxn,
		FromApp: "SUPPLIER",
		ToApp:   supplierReq.ToApp,
		ToId:    "",
		FromId:  "",
		TxnType: "",
		Clock:   nil,
	}

	block := blockchain.Block{
		IsGenesis:     false,
		BlockId:       "",
		ViewType:      "",
		CryptoHash:    "",
		Transaction:   txn,
		InitiatorNode: "",
		Clock:         nil,
	}

	go supplier.RunSupplierContract(block)
	isValid := <-supplier.ContractValid
	if isValid == false {
		t.Errorf("run supplier contract failed")
	}
}

func TestSupplierContractFail(t *testing.T) {
	supplier := application.Supplier{
		ContractValid: make(chan bool),
		MsgChannel:    make(chan *nats.Msg),
	}

	numUnitsToBuy := 5
	numUnitsToBuyWrong := 4
	shippingService := "UPS"
	supplierReq := application.SupplierRequest{
		ToApp:           "",
		NumUnitsToBuy:   numUnitsToBuy,
		AmountPaid:      numUnitsToBuyWrong*application.ManufacturerCostPerUnit + application.Rates[shippingService],
		ShippingService: shippingService,
		ShippingCost:    application.Rates[shippingService],
	}

	jTxn, _ := json.Marshal(&supplierReq)

	txn := &common.Transaction{
		TxnBody: jTxn,
		FromApp: "SUPPLIER",
		ToApp:   supplierReq.ToApp,
		ToId:    "",
		FromId:  "",
		TxnType: "",
		Clock:   nil,
	}

	block := blockchain.Block{
		IsGenesis:     false,
		BlockId:       "",
		ViewType:      "",
		CryptoHash:    "",
		Transaction:   txn,
		InitiatorNode: "",
		Clock:         nil,
	}

	go supplier.RunSupplierContract(block)
	isValid := <-supplier.ContractValid
	if isValid == true {
		t.Errorf("run supplier contract total amount should have failed")
	}

	supplierReq = application.SupplierRequest{
		ToApp:           "",
		NumUnitsToBuy:   numUnitsToBuy,
		AmountPaid:      numUnitsToBuy*application.ManufacturerCostPerUnit + application.Rates[shippingService],
		ShippingService: shippingService,
		ShippingCost:    application.Rates[shippingService] - 1,
	}

	jTxn, _ = json.Marshal(&supplierReq)

	txn = &common.Transaction{
		TxnBody: jTxn,
		FromApp: "SUPPLIER",
		ToApp:   supplierReq.ToApp,
		ToId:    "",
		FromId:  "",
		TxnType: "",
		Clock:   nil,
	}

	block = blockchain.Block{
		IsGenesis:     false,
		BlockId:       "",
		ViewType:      "",
		CryptoHash:    "",
		Transaction:   txn,
		InitiatorNode: "",
		Clock:         nil,
	}

	go supplier.RunSupplierContract(block)
	isValid = <-supplier.ContractValid
	if isValid == true {
		t.Errorf("run supplier contract shipping cost should have failed")
	}
}

func TestBuyerContractSuccess(t *testing.T) {
	buyer := application.Buyer{
		ContractValid: make(chan bool),
		MsgChannel:    make(chan *nats.Msg),
	}

	unitsTransferred := 4
	shippingService := "USPS"

	buyerReq := application.BuyerClientRequest{
		ToApp:            "",
		Type:             "",
		UnitsTransferred: unitsTransferred,
		MoneyTransferred: unitsTransferred*application.SupplierCostPerUnit + application.Rates[shippingService],
		ShippingService:  shippingService,
		ShippingCost:     application.Rates[shippingService],
	}

	jTxn, _ := json.Marshal(&buyerReq)

	txn := &common.Transaction{
		TxnBody: jTxn,
		FromApp: "BUYER",
		ToApp:   buyerReq.ToApp,
		ToId:    "",
		FromId:  "",
		TxnType: "",
		Clock:   nil,
	}

	block := blockchain.Block{
		IsGenesis:     false,
		BlockId:       "",
		ViewType:      "",
		CryptoHash:    "",
		Transaction:   txn,
		InitiatorNode: "",
		Clock:         nil,
	}

	go buyer.RunBuyerContract(block)
	isValid := <-buyer.ContractValid
	if isValid == false {
		t.Errorf("run buyer contract failed")
	}
}

func TestBuyerContractFail(t *testing.T) {
	buyer := application.Buyer{
		ContractValid: make(chan bool),
		MsgChannel:    make(chan *nats.Msg),
	}

	unitsTransferred := 4
	unitsTransferredWrong := 5
	shippingService := "USPS"

	buyerReq := application.BuyerClientRequest{
		ToApp:            "",
		Type:             "",
		UnitsTransferred: unitsTransferred,
		MoneyTransferred: unitsTransferredWrong*application.SupplierCostPerUnit + application.Rates[shippingService],
		ShippingService:  shippingService,
		ShippingCost:     application.Rates[shippingService],
	}

	jTxn, _ := json.Marshal(&buyerReq)

	txn := &common.Transaction{
		TxnBody: jTxn,
		FromApp: "BUYER",
		ToApp:   buyerReq.ToApp,
		ToId:    "",
		FromId:  "",
		TxnType: "",
		Clock:   nil,
	}

	block := blockchain.Block{
		IsGenesis:     false,
		BlockId:       "",
		ViewType:      "",
		CryptoHash:    "",
		Transaction:   txn,
		InitiatorNode: "",
		Clock:         nil,
	}

	go buyer.RunBuyerContract(block)
	isValid := <-buyer.ContractValid
	if isValid == true {
		t.Errorf("run buyer contract total amount should have failed")
	}

	buyerReq = application.BuyerClientRequest{
		ToApp:            "",
		Type:             "",
		UnitsTransferred: unitsTransferred,
		MoneyTransferred: unitsTransferred*application.SupplierCostPerUnit + application.Rates[shippingService],
		ShippingService:  shippingService,
		ShippingCost:     application.Rates[shippingService] + 1,
	}

	jTxn, _ = json.Marshal(&buyerReq)

	txn = &common.Transaction{
		TxnBody: jTxn,
		FromApp: "BUYER",
		ToApp:   buyerReq.ToApp,
		ToId:    "",
		FromId:  "",
		TxnType: "",
		Clock:   nil,
	}

	block = blockchain.Block{
		IsGenesis:     false,
		BlockId:       "",
		ViewType:      "",
		CryptoHash:    "",
		Transaction:   txn,
		InitiatorNode: "",
		Clock:         nil,
	}

	go buyer.RunBuyerContract(block)
	isValid = <-buyer.ContractValid
	if isValid == true {
		t.Errorf("run buyer contract shipping cost should have failed")
	}
}

func TestManufacturerContractSuccess(t *testing.T) {
	manufacturer := application.Manufacturer{
		ContractValid: make(chan bool),
		MsgChannel:    make(chan *nats.Msg),
		IsPrimary:     false,
	}

	numUnitsToSell := 6
	manufacturerReq := application.ManufacturerClientRequest{
		ToApp:               "",
		NumUnitsToSell:      numUnitsToSell,
		AmountToBeCollected: numUnitsToSell * application.ManufacturerCostPerUnit,
	}

	jTxn, _ := json.Marshal(&manufacturerReq)

	txn := &common.Transaction{
		TxnBody: jTxn,
		FromApp: "MANUFACTURER",
		ToApp:   manufacturerReq.ToApp,
		ToId:    "",
		FromId:  "",
		TxnType: "",
		Clock:   nil,
	}

	block := blockchain.Block{
		IsGenesis:     false,
		BlockId:       "",
		ViewType:      "",
		CryptoHash:    "",
		Transaction:   txn,
		InitiatorNode: "",
		Clock:         nil,
	}

	go manufacturer.RunManufacturerContract(block)
	isValid := <-manufacturer.ContractValid
	if isValid == false {
		t.Errorf("run manufacturer contract failed")
	}
}

func TestManufacturerContractFail(t *testing.T) {
	manufacturer := application.Manufacturer{
		ContractValid: make(chan bool),
		MsgChannel:    make(chan *nats.Msg),
		IsPrimary:     false,
	}

	numUnitsToSell := 6
	numUnitsToSellWrong := 5
	manufacturerReq := application.ManufacturerClientRequest{
		ToApp:               "",
		NumUnitsToSell:      numUnitsToSell,
		AmountToBeCollected: numUnitsToSellWrong * application.ManufacturerCostPerUnit,
	}

	jTxn, _ := json.Marshal(&manufacturerReq)

	txn := &common.Transaction{
		TxnBody: jTxn,
		FromApp: "MANUFACTURER",
		ToApp:   manufacturerReq.ToApp,
		ToId:    "",
		FromId:  "",
		TxnType: "",
		Clock:   nil,
	}

	block := blockchain.Block{
		IsGenesis:     false,
		BlockId:       "",
		ViewType:      "",
		CryptoHash:    "",
		Transaction:   txn,
		InitiatorNode: "",
		Clock:         nil,
	}

	go manufacturer.RunManufacturerContract(block)
	isValid := <-manufacturer.ContractValid
	if isValid == true {
		t.Errorf("run manufacturer contract should have failed")
	}
}
