package common

import (
	"strings"
	"sync"

	"github.com/sirupsen/logrus"

	"All-On-Cloud-9/config"
	"bufio"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"os"
)

var (
	GlobalClockLock sync.Mutex
	GlobalClock     = 0
)

func ConfigureLogger(level string) {
	logrus.SetFormatter(&logrus.TextFormatter{
		ForceColors:               true,
		DisableColors:             false,
		ForceQuote:                false,
		DisableQuote:              false,
		EnvironmentOverrideColors: false,
		DisableTimestamp:          false,
		FullTimestamp:             false,
		TimestampFormat:           "",
		DisableSorting:            false,
		SortingFunc:               nil,
		DisableLevelTruncation:    true,
		PadLevelText:              false,
		QuoteEmptyFields:          false,
		FieldMap:                  nil,
		CallerPrettyfier:          nil,
	})
	logrus.SetOutput(os.Stdout)
	switch strings.ToLower(level) {
	case "panic":
		logrus.SetLevel(logrus.PanicLevel)
	case "debug":
		logrus.SetLevel(logrus.DebugLevel)
	case "info":
		logrus.SetLevel(logrus.InfoLevel)
	case "warning", "warn":
		logrus.SetLevel(logrus.WarnLevel)
	}
}

func UpdateGlobalClock(currTimestamp int, local bool) {
	GlobalClockLock.Lock()
	defer GlobalClockLock.Unlock()

	if local {
		GlobalClock += 1
		return
	}
	if currTimestamp > GlobalClock {
		GlobalClock = currTimestamp + 1
	} else {
		GlobalClock += 1
	}
	logrus.WithFields(logrus.Fields{
		"clock": GlobalClock,
	}).Debug("updated the global clock")
}

func checkError(err error) {
	if err != nil {
		fmt.Println("Fatal error ", err.Error())
		os.Exit(1)
	}
}

func FillDigestAndRemoveSign(msg Message) Message {
	_msg := msg
	_msg.Digest = ""
	_msg.PKeySig = ""
	byteMsg, err := json.Marshal(_msg)
	checkError(err)
	digest := sha256.Sum256(byteMsg)
	_msg.Digest = base64.StdEncoding.EncodeToString(digest[:])

	return _msg
}

func savePEMKey(fileName string, key *rsa.PrivateKey) {
	outFile, err := os.Create(fileName)
	checkError(err)
	defer outFile.Close()

	var privateKey = &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	}

	err = pem.Encode(outFile, privateKey)
	checkError(err)
}

func GeneratePrivateKey(appId, id int) {
	reader := rand.Reader
	bitSize := 2048

	key, err := rsa.GenerateKey(reader, bitSize)
	checkError(err)

	savePEMKey(fmt.Sprintf("../config/keys/private%v_%v.pem", appId, id), key)
}

func importPrivateKey(appId, id int) *rsa.PrivateKey {
	privateKeyFile, err := os.Open(fmt.Sprintf("../config/keys/private%v_%v.pem", appId, id))
	checkError(err)

	pemfileinfo, _ := privateKeyFile.Stat()
	var size int64 = pemfileinfo.Size()
	pembytes := make([]byte, size)
	buffer := bufio.NewReader(privateKeyFile)
	_, err = buffer.Read(pembytes)
	data, _ := pem.Decode([]byte(pembytes))
	privateKeyFile.Close()

	privateKeyImported, err := x509.ParsePKCS1PrivateKey(data.Bytes)
	checkError(err)
	return privateKeyImported
}

func importPublicKey(appId, id int) *rsa.PublicKey {
	return &importPrivateKey(appId, id).PublicKey
}

func getNodeId(message Message) int {
	//if message.FromNodeId == "" {
	//	panic("fill FromNodeId")
	//}
	//id, err := strconv.Atoi(message.FromNodeId)
	//if err != nil {
	//	panic(fmt.Sprintf("%q doesn't look like a number.\n", message.FromNodeId))
	//}
	//if id != message.FromNodeNum {
	//	panic(fmt.Sprintf("message.FromNodeNum(%v) is not equal to message.FromNodeId(%v)", message.FromNodeNum, message.FromNodeId))
	//}

	//return id

	return message.FromNodeNum
}

func Encrypt(message Message) Message {
	if message.Digest == "" {
		panic("fill the digest first")
	}
	appId := config.GetAppId(message.FromApp)
	id := getNodeId(message)

	byteDigest, err := base64.StdEncoding.DecodeString(message.Digest)
	checkError(err)
	signature, err := rsa.SignPKCS1v15(rand.Reader, importPrivateKey(appId, id), crypto.SHA256, byteDigest)
	checkError(err)
	message.PKeySig = base64.StdEncoding.EncodeToString(signature)
	return message
}

func Verify(message Message) bool {
	appId := config.GetAppId(message.FromApp)
	id := getNodeId(message)

	byteDigest, err := base64.StdEncoding.DecodeString(message.Digest)
	checkError(err)
	byteSign, err := base64.StdEncoding.DecodeString(message.PKeySig)
	checkError(err)
	err = rsa.VerifyPKCS1v15(importPublicKey(appId, id), crypto.SHA256, byteDigest, byteSign)
	return err == nil && message.Digest == FillDigestAndRemoveSign(message).Digest // Verify sign and validity of digest
}
