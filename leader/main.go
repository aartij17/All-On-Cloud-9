package main

import (
	"All-On-Cloud-9/leader/node"
)

func main() {
	l := node.Leader{0}
	l.HandleReceiveCommand("hello")
}