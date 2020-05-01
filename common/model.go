package common

type Transaction struct {
	LocalXNum  string        `json:"local_transaction_number"`
	GlobalXNum string        `json:"global_transaction_number"`
	Type       string        `json:"transaction_type"` // local/global transaction type
	TxnId      string        `json:"txn_id"`
	ToId       string        `json:"to_id"`
	FromId     string        `json:"from_id"`
	CryptoHash string        `json:"crypto_hash"`
	TxnType    string        `json:"transaction_type"`
	Clock      *LamportClock `json:"lamport_clock"`
}

type Message struct {
	MessageType string       `json:"message_type"`
	Timestamp   int          `json:"client_timestamp,omitempty"`
	FromNodeId  string       `json:"from_node_id"`
	FromNodeNum int          `json:"from_node_num"`
	Txn         *Transaction `json:"transaction,omitempty"`
	Digest      string       `json:"digest"`
	PKeySig     string       `json:"pkey_sig"`
}

// LamportClock is used for ordering the local/global transactions
type LamportClock struct {
	PID   int `json:"pid"`
	Clock int `json:"clock"`
}
