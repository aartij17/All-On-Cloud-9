package application

import (
	"All-On-Cloud-9/server/blockchain"
	"encoding/json"

	log "github.com/sirupsen/logrus"
)

// This file has all the contracts that are invoked when the consensus is reached and just BEFORE
// the transactions are executed (in our case, the transactions are additions of the blocks to the blockchain
func RunBuyerContract(block *blockchain.Block, contractValid chan bool) {
	// business logic here
	// based on the results, send a true/false value to the channel
	var err error
	bTxn := BuyerClientRequest{}
	err = json.Unmarshal(block.Transaction.TxnBody, &bTxn)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Error("error runBuyerContract")
		contractValid <- false
		return
	}

	if bTxn.UnitsTransferred*SupplierCostPerUnit+bTxn.ShippingCost == bTxn.MoneyTransferred {
		// Verify that the shipping cost is correct for the given shipping service
		if _, ok := Rates[bTxn.ShippingService]; ok {
			if Rates[bTxn.ShippingService] == bTxn.ShippingCost {
				contractValid <- true
				return
			}
		}
	}
	contractValid <- false
}

func RunManufacturerContract(block *blockchain.Block, contractValid chan bool) {
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
		contractValid <- false
		return
	}

	if mTxn.NumUnitsToSell*ManufacturerCostPerUnit == mTxn.AmountToBeCollected {
		log.Info("MANUFACTURER Smart contract VALID")
		contractValid <- true
		return
	}
	log.Info("MANUFACTURER Smart contract INVALID")
	// based on the results, send a true/false value to the channel
	contractValid <- false
}

func RunSupplierContract(block *blockchain.Block, contractValid chan bool) {
	// business logic here
	// The supplier needs to purchase from the
	var err error

	sTxn := SupplierRequest{}
	err = json.Unmarshal(block.Transaction.TxnBody, &sTxn)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Error("error runSupplierContract")
		contractValid <- false
		return
	}
	// Verify the amount needed is correct
	if sTxn.NumUnitsToBuy*ManufacturerCostPerUnit+sTxn.ShippingCost == sTxn.AmountPaid {

		// Verify that the shipping cost is correct for the given shipping service
		if _, ok := Rates[sTxn.ShippingService]; ok {
			if Rates[sTxn.ShippingService] == sTxn.ShippingCost {
				contractValid <- true
				return
			}
		}
	}

	// based on the results, send a true/false value to the channel
	contractValid <- false
}

func RunCarrierContract(block *blockchain.Block, contractValid chan bool) {
	// business logic: Carrier asks another process for a fee for its services
	// based on the results, send a true/false value to the channel
	var err error
	cTxn := CarrierClientRequest{}
	err = json.Unmarshal(block.Transaction.TxnBody, &cTxn)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Error("error runCarrierContract")
		contractValid <- false
		return
	}

	// Check that the fee requested is nonzero
	if cTxn.Fee > 0 {
		contractValid <- true
		return
	}
	contractValid <- false
}
