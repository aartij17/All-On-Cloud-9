package config

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"os"

	log "github.com/Sirupsen/logrus"
)

const (
	APP1 = 1
	APP2 = 2
	APP3 = 3

	NODE11 = 1
	NODE12 = 2
	NODE13 = 3

	NODE21 = 1
	NODE22 = 2
	NODE23 = 3

	NODE31 = 1
	NODE32 = 2
	NODE33 = 3

	// ORDERER nodes which are NOT part of the agents serving the applications
	ORDERER1 = 1
	ORDERER2 = 2
	ORDERER3 = 3

	//ORDERER21 = "ORDERER21"
	//ORDERER22 = "ORDERER22"
	//ORDERER23 = "ORDERER23"
	//
	//ORDERER31 = "ORDERER31"
	//ORDERER32 = "ORDERER32"
	//ORDERER33 = "ORDERER33"
)

var (
	APP1_NODES = [...]int{NODE11, NODE12, NODE13}
	APP2_NODES = [...]int{NODE21, NODE22, NODE23}
	APP3_NODES = [...]int{NODE31, NODE32, NODE33}

	APP_ORDERERS = [...]int{ORDERER1, ORDERER2, ORDERER3}
	//APP2_ORDERERS = [...]string{ORDERER21, ORDERER22, ORDERER23}
	//APP3_ORDEERES = [...]string{ORDERER31, ORDERER32, ORDERER33}

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

func LoadConfig(ctx context.Context, filepath string) {
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
