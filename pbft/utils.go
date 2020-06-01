package pbft

import (
	"All-On-Cloud-9/common"
	"All-On-Cloud-9/config"

	log "github.com/Sirupsen/logrus"
)

func PipeInHierarchicalLocalConsensus(pbftNode *PbftNode) {
	if config.SystemConfig.GlobalConsensusAlgo != common.GLOBAL_CONSENSUS_ALGO_HEIRARCHICAL &&
		config.SystemConfig.GlobalConsensusAlgo != common.GLOBAL_CONSENSUS_ALGO_TH_HEIRARCHICAL {
		log.Info("not gonna start heirarchical consensus listener, since the config doesn't permit it")
		return
	}
	log.Info("starting heirarchical consensus listener")
	for {
		txn := <-pbftNode.LocalConsensusRequired
		pbftNode.MessageIn <- txn
	}
}
