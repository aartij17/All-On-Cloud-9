package pbft

func PipeInLocalConsensus(pbftNode *PbftNode) {
	for {
		txn := <-pbftNode.LocalConsensusRequired
		pbftNode.MessageIn <- txn
	}
}
