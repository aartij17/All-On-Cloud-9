package config

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	log "github.com/Sirupsen/logrus"
)

const (
	APP_MANUFACTURER = "APP_MANUFACTURER"
	APP_SUPPLIER     = "APP_SUPPLIER"
	APP_BUYER        = "APP_BUYER"
	APP_CARRIER      = "APP_CARRIER"

	MA_NODE = "MA_NODE_%d"
	S_NODE  = "S_NODE_%d"
	B_NODE  = "B_NODE_%d"
	C_NODE  = "C_NODE_%d"

	// ORDERER nodes which are NOT part of the agents serving the applications
	ORDERER1 = 1
	ORDERER2 = 2
	ORDERER3 = 3
)

var (
	MANUFACTURER_NODES []string
	SUPPLIER_NODES     []string
	BUYER_NODES        []string
	CARRIER_NODES      []string

	APP_ORDERERS = [...]int{ORDERER1, ORDERER2, ORDERER3}

	SystemConfig *Config
)

type Servers struct {
	Host string `json:"host"`
	Port int    `json:"port"`
}

type Applications struct {
	Application1 *ApplicationInstance `json:"APPLICATION_1,omitempty"`
	Application2 *ApplicationInstance `json:"APPLICATION_2,omitempty"`
	Application3 *ApplicationInstance `json:"APPLICATION_3,omitempty"`
}

type ApplicationInstance struct {
	Servers []*Servers `json:"servers"`
}

type Orderers struct {
	Servers []*Servers `json:"servers"`
}

type NatsServers struct {
	Servers []string `json:"servers"`
}

type Config struct {
	AppInstance *Applications `json:"application_instance"`
	Orderers    *Orderers     `json:"orderers"`
	Nats        *NatsServers  `json:"nats"`
}

func initNodeIds() {
	for i := 0; i < 5; i++ {
		MANUFACTURER_NODES = append(MANUFACTURER_NODES, fmt.Sprintf(MA_NODE, i))
		SUPPLIER_NODES = append(SUPPLIER_NODES, fmt.Sprintf(S_NODE, i))
		BUYER_NODES = append(BUYER_NODES, fmt.Sprintf(B_NODE, i))
		CARRIER_NODES = append(CARRIER_NODES, fmt.Sprintf(C_NODE, i))
	}
	log.WithFields(log.Fields{
		"manufacturer": MANUFACTURER_NODES,
		"supplier":     SUPPLIER_NODES,
		"buyer":        BUYER_NODES,
		"carrier":      CARRIER_NODES,
	}).Info("initialized all app nodes with their app IDs")
}

func LoadConfig(ctx context.Context, filepath string) {
	initNodeIds()
	jsonFile, err := os.Open(filepath)
	if err != nil {
		log.WithFields(log.Fields{
			"err":  err.Error(),
			"path": filepath,
		}).Error("error opening config file")
	}
	file, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err.Error(),
		}).Error("error reading config file")
	}
	err = json.Unmarshal(file, &SystemConfig)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err.Error(),
		}).Error("error unmarshalling config into the conf object")
	}
}

//// STUB: For testing only
//func main() {
//	LoadConfig(nil, "/Users/aartij17/go/src/All-On-Cloud-9/config/config.json")
//	fmt.Println(SystemConfig.Nats.Servers)
//}
