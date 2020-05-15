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

func NewPbftNode(ctx context.Context, nc *nats.Conn, application string, failureTolerance int, totalNodes int, id int) *PbftNode {
	node := newPbftNode(ctx, nc, application, failureTolerance, totalNodes, id)
	node.subToNatsChannels(application)
	go node.startMessageListeners(node.msgChannel)
	return node
}