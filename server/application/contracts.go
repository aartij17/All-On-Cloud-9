package application

import (
	"All-On-Cloud-9/server/blockchain"
	"encoding/json"

	log "github.com/Sirupsen/logrus"
)

// This file has all the contracts that are invoked when the consensus is reached and just BEFORE
// the transactions are executed (in our case, the transactions are additions of the blocks to the blockchain
func (buyer *Buyer) RunBuyerContract(block *blockchain.Block) {
	// business logic here
	// based on the results, send a true/false value to the channel
	var err error
	bTxn := BuyerClientRequest{}
	err = json.Unmarshal(block.Transaction.TxnBody, &bTxn)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Error("error runBuyerContract")
		buyer.ContractValid <- false
		return
	}

	if bTxn.UnitsTransferred*SupplierCostPerUnit+bTxn.ShippingCost == bTxn.MoneyTransferred {
		// Verify that the shipping cost is correct for the given shipping service
		if _, ok := Rates[bTxn.ShippingService]; ok {
			if Rates[bTxn.ShippingService] == bTxn.ShippingCost {
				buyer.ContractValid <- true
				return
			}
		}
	}

	buyer.ContractValid <- false
}

func (manufacturer *Manufacturer) RunManufacturerContract(block *blockchain.Block) {
	// business logic here
	// The manufacturer block will advertise how many units it has and
	// how much money everything costs to another application, most likely the seller
	var err error
	mTxn := ManufacturerClientRequest{}
	err = json.Unmarshal(block.Transaction.TxnBody, &mTxn)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Error("error runManufacturerContract")
		manufacturer.ContractValid <- false
		return
	}

	if mTxn.NumUnitsToSell*ManufacturerCostPerUnit == mTxn.AmountToBeCollected {
		log.Info("MANUFACTURER Smart contract VALID")
		manufacturer.ContractValid <- true
		return
	}
	log.Info("MANUFACTURER Smart contract INVALID")
	// based on the results, send a true/false value to the channel
	manufacturer.ContractValid <- false
}

func (supplier *Supplier) RunSupplierContract(block *blockchain.Block) {
	// business logic here
	// The supplier needs to purchase from the
	var err error

	sTxn := SupplierRequest{}
	err = json.Unmarshal(block.Transaction.TxnBody, &sTxn)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Error("error runSupplierContract")
		supplier.ContractValid <- false
		return
	}
	// Verify the amount needed is correct
	if sTxn.NumUnitsToBuy*ManufacturerCostPerUnit+sTxn.ShippingCost == sTxn.AmountPaid {

		// Verify that the shipping cost is correct for the given shipping service
		if _, ok := Rates[sTxn.ShippingService]; ok {
			if Rates[sTxn.ShippingService] == sTxn.ShippingCost {
				supplier.ContractValid <- true
				return
			}
		}
	}

	// based on the results, send a true/false value to the channel
	supplier.ContractValid <- false
}

func (carrier *Carrier) RunCarrierContract(block *blockchain.Block) {
	// business logic: Carrier asks another process for a fee for its services
	// based on the results, send a true/false value to the channel
	var err error
	cTxn := CarrierClientRequest{}
	err = json.Unmarshal(block.Transaction.TxnBody, &cTxn)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Error("error runCarrierContract")
		carrier.ContractValid <- false
		return
	}

	// Check that the fee requested is nonzero
	if cTxn.Fee > 0 {
		carrier.ContractValid <- true
		return
	}

	carrier.ContractValid <- false
}
