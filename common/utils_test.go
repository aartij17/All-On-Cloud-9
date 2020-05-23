package common

import (
	"strconv"
	"testing"

	log "github.com/Sirupsen/logrus"
)

func checkMessage(message Message, t *testing.T) {
	if !Verify(Encrypt(FillDigestAndRemoveSign(message))) {
		t.Error("Signature not verified: ")
		t.Error(message)
		t.Error(FillDigestAndRemoveSign(message))
		t.Error(Encrypt(FillDigestAndRemoveSign(message)))
	}
	log.WithField("msg", Encrypt(FillDigestAndRemoveSign(message))).Info("verified")
}

func TestEncrypt(t *testing.T) {
	message := Message{
		ToApp:       "",
		MessageType: "",
		Timestamp:   0,
		Txn:         nil,
		Digest:      "",
		PKeySig:     "",
	}

	var messages [4][4]Message
	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			message.FromApp = strconv.Itoa(i)
			message.FromNodeId = strconv.Itoa(j)
			message.FromNodeNum = j

			messages[i][j] = message
		}
	}

	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			checkMessage(messages[i][j], t)
		}
	}
}
