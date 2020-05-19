package common

import (
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
	"strconv"
)

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

func getAppId(message Message) int {
	if message.FromApp == "" {
		panic("fill FromApp")
	}
	appId, err := strconv.Atoi(message.FromApp)
	if err != nil {
		appId = config.GetAppId(message.FromApp)
	}

	return appId
}

func getNodeId(message Message) int {
	if message.FromNodeId == "" {
		panic("fill FromNodeId")
	}
	id, err := strconv.Atoi(message.FromNodeId)
	if err != nil {
		panic(fmt.Sprintf("%q doesn't look like a number.\n", message.FromNodeId))
	}
	if id != message.FromNodeNum {
		panic(fmt.Sprintf("message.FromNodeNum(%v) is not equal to message.FromNodeId(%v)", message.FromNodeNum, message.FromNodeId))
	}

	return id
}

func Encrypt(message Message) Message {
	if message.Digest == "" {
		panic("fill the digest first")
	}
	appId := getAppId(message)
	id := getNodeId(message)

	byteDigest, err := base64.StdEncoding.DecodeString(message.Digest)
	checkError(err)
	signature, err := rsa.SignPKCS1v15(rand.Reader, importPrivateKey(appId, id), crypto.SHA256, byteDigest)
	checkError(err)
	message.PKeySig = base64.StdEncoding.EncodeToString(signature)
	return message
}

func Verify(message Message) bool {
	appId := getAppId(message)
	id := getNodeId(message)

	byteDigest, err := base64.StdEncoding.DecodeString(message.Digest)
	checkError(err)
	byteSign, err := base64.StdEncoding.DecodeString(message.PKeySig)
	checkError(err)
	err = rsa.VerifyPKCS1v15(importPublicKey(appId, id), crypto.SHA256, byteDigest, byteSign)
	return err == nil && message.Digest == FillDigestAndRemoveSign(message).Digest // Verify sign and validity of digest
}