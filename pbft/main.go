package pbft

import (
	"All-On-Cloud-9/messenger"
	"context"
	"github.com/nats-io/nats.go"
	log "github.com/sirupsen/logrus"
	"os"
	"strings"
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

func (node *PbftNode) subToInterAppNats(ctx context.Context, nc *nats.Conn, suffix string) {
	var (
		err error
	)
	err = messenger.SubscribeToInbox(ctx, nc, NATS_PBFT_INBOX + suffix, node.MsgChannel)
	if err != nil {
		log.WithFields(log.Fields{
			"error":       err.Error(),
			"application": APPLICATION,
			"topic":       NATS_PBFT_INBOX + suffix,
		}).Error("error subscribing to the nats topic")
	}
}

func setupPbftNode(ctx context.Context, nc *nats.Conn, suffix string) *PbftNode {
	node := NewPbftNode()
	node.subToInterAppNats(ctx, nc, suffix)
	go node.startInterAppNatsListener(ctx, node.MsgChannel)
	return node
}
