#!/bin/bash

# Base case:
# 1 leader, 50 proposers, 50 consensus nodes, 50 replicas
go run main.go -nodetype 0  & # Leader
go run main.go -nodetype 1 -nodeId 0  &# Proposer
go run main.go -nodetype 2 & # Consensus
go run main.go -nodetype 2  & # Consensus
go run main.go -nodetype 2  &# Consensus
go run main.go -nodetype 3  &

echo "DONE"
# Flood leader with requests


# Get performance throughput


# Scale up to:
# 1 leader, 100 proposers, 50 consensus nodes, 50 replicas


# Flood leader with same number of requests


# Get performance throughput

