package common

type Transaction struct {
	TxnId      string        `json:"txn_id"`
	ToId       string        `json:"to_id"`
	FromId     string        `json:"from_id"`
	CryptoHash string        `json:"crypto_hash"`
	Amount     int           `json:"amount"`
	TxnType    string        `json:"transaction_type"`
	Clock      *LamportClock `json:"lamport_clock"`
}

type Message struct {
	Txn     *Transaction `json:"transaction"`
	Digest  string       `json:"digest"`
	PKeySig string       `json:"pkey_sig"`
}

// LamportClock is used for ordering the local/global transactions
type LamportClock struct {
	PID   int `json:"pid"`
	Clock int `json:"clock"`
}
