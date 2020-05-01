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
)

var (
	NATS_ORDERER_SUBJECTS = [...]string{NATS_ORD_ORDER, NATS_ORD_SYNC}
)
