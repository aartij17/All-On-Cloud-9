package pbft

import (
	"context"
	"os"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/nats-io/nats.go"
)

func _inbox(suffix string) string {
	if suffix == GLOBAL_APPLICATION {
		return GLOBAL_APPLICATION
	}
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
	node.initTimer(globalState, node.generateGlobalBroadcast(localState))
	node.subToNatsChannels(application)
	go node.startMessageListeners(node.msgChannel)
	go node.handleLocalOut(localState)
	go node.handleGlobalOut(globalState)
	return node
}
