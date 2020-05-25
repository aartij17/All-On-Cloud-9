package pbft

func PipeInHierarchicalLocalConsensus(pbftNode *PbftNode) {
	for {
		txn := <-pbftNode.LocalConsensusRequired
		pbftNode.MessageIn <- txn
	}
}
