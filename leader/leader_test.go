package main

import (
	"testing"
	"All-On-Cloud-9/leader/node"
	"All-On-Cloud-9/common"
	"reflect"
	"fmt"
)

func TestUnion1(t *testing.T) {
	// Test that all dependencies are unioned together
	v0 := common.Vertex{0,0}
	v1 := common.Vertex{0,1}
	v2 := common.Vertex{0,2}
	v3 := common.Vertex{1,1}
	v4 := common.Vertex{1,2}
	v5 := common.Vertex{2,1}
	v6 := common.Vertex{2,2}

	message1 := common.MessageEvent{&v0, "Hello", []*common.Vertex{&v1, &v2}}
	message2 := common.MessageEvent{&v0, "Hello", []*common.Vertex{&v3, &v4}}
	message3 := common.MessageEvent{&v0, "Hello", []*common.Vertex{&v5, &v6}}

	messages := []*common.MessageEvent{&message1, &message2, &message3}
	newMessage := node.Union(messages)
	
	// Now check that newMessages has been created correctly
	if !reflect.DeepEqual(newMessage.VertexId, v0) {
		t.Errorf("newMessage VertexId = (%d, %d); want (%d, %d)", 
		newMessage.VertexId.Index, newMessage.VertexId.Id, 
		v0.Index, v0.Id)
	}

	if newMessage.Message != "Hello" {
		t.Errorf("newMessage Message = %s; want \"Hello\"", newMessage.Message)
	}

	for idx, item := range newMessage.Deps {
		if !reflect.DeepEqual(messages[idx], item) {
			t.Errorf("newMessage Dep[%d] = (%d, %d); want (%d, %d)", idx,
			item.Index, item.Id, 
			messages[idx].VertexId.Index, messages[idx].VertexId.Id)
		}
	}

	fmt.Println("Union1 test SUCCESS")
}

// func TestUnion2(t *testing.T) {
// 	// Test that duplicate dependencies are not unioned Twice
// 	messages := make([]*common.MessageEvent, 5)
	
// }