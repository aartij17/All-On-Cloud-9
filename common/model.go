package common

type MessageEvent struct {
	VertexId *Vertex   `json:"vertex"`
	Message  []byte    `json:"message"`
	Deps     []*Vertex `json:"dependency,omitempty"`
}

type Vertex struct {
	Index int `json:"index"`
	Id    int `json:"id"`
}

type ConsensusMessage struct {
	VertexId   *Vertex `json:"vertex"`
	Release    int     `json:"release"`
	ProposerId int     `json:"proposerId"`
}

type Transaction struct {
	TxnBody []byte        `json:"txn_body"`
	FromApp string        `json:"from_app"`
	ToApp   string        `json:"to_app"`
	ToId    string        `json:"to_id,omitempty"`
	FromId  string        `json:"from_id,omitempty"`
	TxnType string        `json:"transaction_type"`
	Clock   *LamportClock `json:"lamport_clock"`
}

type Message struct {
	ToApp       string        `json:"to_application"`
	FromApp     string        `json:"from_application"`
	MessageType string        `json:"message_type"`
	Timestamp   int           `json:"client_timestamp,omitempty"`
	FromNodeId  string        `json:"from_node_id"`
	FromNodeNum int           `json:"from_node_num"`
	Txn         *Transaction  `json:"transaction,omitempty"`
	Digest      string        `json:"digest"`
	PKeySig     string        `json:"pkey_sig"`
	Clock       *LamportClock `json:"lamport_clock"`
}

// LamportClock is used for ordering the local/global transactions
type LamportClock struct {
	PID   string `json:"pid"`
	Clock int    `json:"clock"`
}
