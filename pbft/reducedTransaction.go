package pbft

import (
	"All-On-Cloud-9/common"
	"fmt"
)

type reducedTransaction struct {
	TxnBody string               `json:"txn_body"`
	FromApp string               `json:"from_app"`
	ToApp   string               `json:"to_app"`
	ToId    string               `json:"to_id,omitempty"`
	FromId  string               `json:"from_id,omitempty"`
	TxnType string               `json:"transaction_type"`
	Clock   *common.LamportClock `json:"lamport_clock"`
}

func newReducedTransaction(txn common.Transaction) reducedTransaction {
	return reducedTransaction{
		TxnBody: fmt.Sprintf("%q", txn.TxnBody),
		FromApp: txn.FromApp,
		ToApp:   txn.ToApp,
		ToId:    txn.ToId,
		FromId:  txn.FromId,
		TxnType: txn.TxnType,
		Clock:   txn.Clock,
	}
}
