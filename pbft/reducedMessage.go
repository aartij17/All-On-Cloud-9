package pbft

import "All-On-Cloud-9/common"

type reducedMessage struct {
	messageType string
	Txn         *common.Transaction
}
