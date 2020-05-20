package pbft

type reducedMessage struct {
	messageType string
	Txn         reducedTransaction
	appId int
}
