package main

import (
	"testing"
	"All-On-Cloud-9/dependency/node"
	"All-On-Cloud-9/common"
)

func TestComputeConflicts(t *testing.T) {
	v0 := common.Vertex{0,0}
	message1 := common.MessageEvent{&v0, "Hello", []*common.Vertex{&v0}}

	depService := depsnode.DepsServiceNode{}
	_ = depService.ComputeConflictingMessages(&message1)
	t.Errorf("STUB: This functon is still unimplemented")
}