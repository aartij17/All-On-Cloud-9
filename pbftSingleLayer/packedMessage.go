package pbftSingleLayer

import "All-On-Cloud-9/common"

type packedMessage struct {
	Msg common.Message     `json:"msg"`
	Txn common.Transaction `json:"txn"`
}
