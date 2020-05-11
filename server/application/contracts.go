package application

import (
	"All-On-Cloud-9/server/blockchain"
	"All-On-Cloud-9/common"
	"encoding/json"
	log "github.com/Sirupsen/logrus"
)

// This file has all the contracts that are invoked when the consensus is reached and just BEFORE
// the transactions are executed (in our case, the transactions are additions of the blocks to the blockchain
func (buyer *Buyer) runBuyerContract() {
	// business logic here
	// based on the results, send a true/false value to the channel
	buyer.ContractValid <- true
}

func (manufacturer *Manufacturer) runManufacturerContract() {
	// business logic here
	// based on the results, send a true/false value to the channel
	manufacturer.ContractValid <- true
}

func (supplier *Supplier) runSupplierContract(block blockchain.Block, MessageType string) {
	// business logic here
	var err error
	// There are two types of Supplier transactions: Buy and Sell
	// (These rules are kinda arbitrary. I made them up as I go)
	// For Buy requests, each UnitToBuy cost is $5
	// For Sell requests, The amount paid must include the shipping cost 
	// (determined by shipping service)
	switch MessageType {
	case common.BUY_REQUEST_TYPE:
		sTxn := SupplierBuyRequest{}
		err = json.Unmarshal(block.Transaction.TxnBody, &sTxn)
		if err != nil {
			log.WithFields(log.Fields{
				"error": err.Error(),
			}).Error("error runSupplierContract: BUY_REQUEST_TYPE")
		}

		if sTxn.NumUnitsToBuy * SUPPLIER_COST_PER_UNIT == sTxn.AmountPaid {
			supplier.ContractValid <- true
			return
		}
		
	case common.SELL_REQUEST_TYPE:
		sTxn := SupplierSellRequest{}
		err = json.Unmarshal(block.Transaction.TxnBody, &sTxn)
		if err != nil {
			log.WithFields(log.Fields{
				"error": err.Error(),
			}).Error("error runSupplierContract: SELL_REQUEST_TYPE")
		}

		if sTxn.NumUnitsToSell * SUPPLIER_COST_PER_UNIT + sTxn.ShippingCost == sTxn.AmountPaid {
			// Verify that the shipping cost is correct for the given shipping service
			if val, ok := rates[sTxn.ShippingService]; ok {
				if rates[sTxn.ShippingService] == sTxn.ShippingCost {
					supplier.ContractValid <- true
					return
				}
			}
		}

	}

	// based on the results, send a true/false value to the channel
	supplier.ContractValid <- false
}

func (carrier *Carrier) runCarrierContract() {
	// business logic here
	// based on the results, send a true/false value to the channel
	carrier.ContractValid <- true
}
