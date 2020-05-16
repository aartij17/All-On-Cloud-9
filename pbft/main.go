package pbft

import (
	"context"
	log "github.com/Sirupsen/logrus"
	"github.com/nats-io/nats.go"
	"os"
	"strings"
)

func _inbox(suffix string) string {
	return NATS_PBFT_INBOX + "_" + suffix
}

func configureLogger(level string) {
	log.SetOutput(os.Stderr)
	switch strings.ToLower(level) {
	case "panic":
		log.SetLevel(log.PanicLevel)
	case "debug":
		log.SetLevel(log.DebugLevel)
	case "info":
		log.SetLevel(log.InfoLevel)
	case "warning", "warn":
		log.SetLevel(log.WarnLevel)
	}
}

func NewPbftNode(
	ctx context.Context,
	nc *nats.Conn,
	application string,
	failureTolerance int,
	totalNodes int,
	globalFailureTolerance int,
	globalTotalNodes int,
	id int,
	appId int,
) *PbftNode {
	localState := newPbftState(failureTolerance, totalNodes, application)
	globalState := newPbftState(globalFailureTolerance, globalTotalNodes, GLOBAL_APPLICATION)
	node := newPbftNode(ctx, nc, id, appId, localState, globalState)
	node.initTimer(localState, node.generateLocalBroadcast())
	//node.initTimer(globalState, )
	node.subToNatsChannels(application)
	go node.startMessageListeners(node.msgChannel)
	return node
}
