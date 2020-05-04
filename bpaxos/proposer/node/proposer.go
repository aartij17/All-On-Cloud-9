package proposer

import (
	"All-On-Cloud-9/common"
	"fmt"
)

type Proposer struct {
}

func (proposer *Proposer) HandleReceive(message *common.MessageEvent) {
	proposer.SendResult(message)
}

func (proposer *Proposer) SendResult(message *common.MessageEvent) {
	fmt.Println("send consensus result")
}
