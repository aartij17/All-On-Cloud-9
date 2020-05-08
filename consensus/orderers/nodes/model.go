package nodes

import "All-On-Cloud-9/common"

type Message struct {
	MessageType   string              `json:"message_type"`
	Timestamp     int                 `json:"client_timestamp,omitempty"`
	Transaction   *common.Transaction `json:"transaction,omitempty"`
	Digest        string              `json:"message_digest,omitempty"`
	Hash          string              `json:"crypto_hash,omitempty"`
	FromNodeId    string              `json:"from_node_id"`
	FromNodeNum   int                 `json:"from_node_num"`
	IsFromPrimary bool                `json:"is_from_primary,omitempty"`
}
