package pbft

import (
	"context"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/nats-io/nats.go"
)

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
	id int,
	appId int,
) *PbftNode {
	globalState := newPbftState(GLOBAL_APPLICATION)
	node := newPbftNode(ctx, nc, id, appId, globalState)
	node.initTimer(globalState, node.generateGlobalBroadcast())
	node.subToNatsChannels(application)
	go node.startMessageListeners(node.msgChannel)
	go node.handleGlobalOut(globalState)
	return node
}
