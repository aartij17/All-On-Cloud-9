package main

import (
	"testing"
	"All-On-Cloud-9/leader/node"
	"All-On-Cloud-9/common"
	"reflect"
)

func TestUnion1(t *testing.T) {
	// Test that all dependencies are unioned together
	v0 := Vertex{0,0}
	v1 := Vertex{0,1}
	v2 := Vertex{0,2}
	v3 := Vertex{1,1}
	v4 := Vertex{1,2}
	v5 := Vertex{2,1}
	v6 := Vertex{2,2}

	message1 := MessageEvent{v0, "Hello", []*Vertex{v1, v2}}
	message2 := MessageEvent{v0, "Hello", []*Vertex{v3, v4}}
	message3 := MessageEvent{v0, "Hello", []*Vertex{v5, v6}}

	messages := []*MessageEvent{message1, message2, message3}
	newMessage := Union(messages)
	
}

func TestUnion2(t *testing.T) {
	// Test that duplicate dependencies are not unioned Twice
	messages := make([]*MessageEvent, 5)
	
}