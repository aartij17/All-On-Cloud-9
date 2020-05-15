package application

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

func (supplier *Supplier) runSupplierContract() {
	// business logic here
	// based on the results, send a true/false value to the channel
	supplier.ContractValid <- true
}

func (carrier *Carrier) runCarrierContract() {
	// business logic here
	// based on the results, send a true/false value to the channel
	carrier.ContractValid <- true
}
