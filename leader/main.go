package main

import (
	"All-On-Cloud-9/leader/node"
)

func main() {
	l := leadernode.Leader{0}
	l.HandleReceiveCommand("hello")
}