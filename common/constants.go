package common

const (
	INTERNAL_TXN  = "INTERNAL_TXN"
	CROSS_APP_TXN = "CROSS_APPLICATION_TXN"

	LOCAL_TXN_NUM  = "LTXN-%d%d"
	GLOBAL_TXN_NUM = "GTXN-%d%d-%d"

	// Orderer Message Types
	O_REQUEST = "REQUEST"
	O_ORDER   = "ORDER"
	O_SYNC    = "SYNC"

	// -------------- inter application messages --------------
	// Message from primary agent of the sender application to the receiver application
	NATS_ORD_REQUEST = "NATS_ORDERER_REQUEST"
	NATS_APPS_TXN    = "NATS_APP_TXN"

	// NATS inbox messages
	// ORDERER MESSAGES
	NATS_ORD_ORDER = "NATS_ORDERER_ORDER"
	NATS_ORD_SYNC  = "NATS_ORDERER_SYNC"

	// NATS start/stop messages
	NATS_CONSENSUS_INITIATE_MSG = "NATS_CONSENSUS_START"
	NATS_CONSENSUS_DONE         = "NATS_CONSENSUS_DONE"

	LEADER_TO_DEPS        = "LEADER_TO_DEPS"
	DEPS_TO_LEADER        = "DEPS_TO_LEADER"
	LEADER_TO_PROPOSER    = "LEADER_TO_PROPOSER"
	PROPOSER_TO_CONSENSUS = "PROPOSER_TO_CONSENSUS"
	CONSENSUS_TO_PROPOSER = "CONSENSUS_TO_PROPOSER"
	PROPOSER_TO_REPLICA   = "PROPOSER_TO_REPLICA"
	CLIENT_TO_LEADER      = "CLIENT_TO_LEADER"

	// Number of tolerable failures
	F = 1

	// Number of BPaxos Proposer nodes
	NUM_PROPOSERS                  = 1
	CONSENSUS_TIMEOUT_MILLISECONDS = 1000
)

var (
	NATS_ORDERER_SUBJECTS = [...]string{NATS_ORD_ORDER, NATS_ORD_SYNC}
)
