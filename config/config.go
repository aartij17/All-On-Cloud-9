package config

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"

	log "github.com/Sirupsen/logrus"
)

const (
	APP_MANUFACTURER = "MANUFACTURER"
	APP_SUPPLIER     = "SUPPLIER"
	APP_BUYER        = "BUYER"
	APP_CARRIER      = "CARRIER"

	NODE_NAME = "%s_%d"

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
	AppManufacturer *ApplicationInstance `json:"MANUFACTURER,omitempty"`
	AppBuyer        *ApplicationInstance `json:"BUYER,omitempty"`
	AppSupplier     *ApplicationInstance `json:"SUPPLIER,omitempty"`
	AppCarrier      *ApplicationInstance `json:"CARRIER,omitempty"`
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

func getAppNum(appName string) int {
	switch appName {
	case APP_BUYER:
		return 0
	case APP_CARRIER:
		return 1
	case APP_MANUFACTURER:
		return 2
	case APP_SUPPLIER:
		return 3
	}
	panic("no such app: " + appName)
}

func GetAppCnt() int {
	return 4
}

func GetAppId(appName string) int {
	if appName == "" {
		panic("fill FromApp")
	}
	appId, err := strconv.Atoi(appName)
	if err != nil {
		appId = getAppNum(appName)
	}

	return appId
}

func GetAppNodeCnt(appName string) int {
	appId := GetAppId(appName)
	switch appId {
	case 0:
		return len(SystemConfig.AppInstance.AppBuyer.Servers)
	case 1:
		return len(SystemConfig.AppInstance.AppCarrier.Servers)
	case 2:
		return len(SystemConfig.AppInstance.AppManufacturer.Servers)
	case 3:
		return len(SystemConfig.AppInstance.AppSupplier.Servers)
	}

	panic("no such app: " + appName)
}

func initNodeIds() {
	for i := 0; i < 5; i++ {
		MANUFACTURER_NODES = append(MANUFACTURER_NODES, fmt.Sprintf(NODE_NAME, APP_MANUFACTURER, i))
		SUPPLIER_NODES = append(SUPPLIER_NODES, fmt.Sprintf(NODE_NAME, APP_SUPPLIER, i))
		BUYER_NODES = append(BUYER_NODES, fmt.Sprintf(NODE_NAME, APP_BUYER, i))
		CARRIER_NODES = append(CARRIER_NODES, fmt.Sprintf(NODE_NAME, APP_CARRIER, i))
	}
	log.WithFields(log.Fields{
		"manufacturer": MANUFACTURER_NODES,
		"supplier":     SUPPLIER_NODES,
		"buyer":        BUYER_NODES,
		"carrier":      CARRIER_NODES,
	}).Info("initialized all app nodes with their app IDs")
}

func LoadConfig(ctx context.Context, filepath string) {
	//initNodeIds()
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
//}
