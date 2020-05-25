#!/bin/bash

rm -f examine.txt
rm -f statuslog.txt
# Base case:
# 1 leader, 50 proposers, 50 consensus nodes, 50 replicas
/usr/local/go/bin/go run main.go -nodetype 0  -numberProps 50 > examine.txt & # Leader

for j in `seq 0 49`;
do
    /usr/local/go/bin/go run main.go -nodetype 1 -nodeId $j & # Proposer
done

for i in `seq 0 49`;
do
   /usr/local/go/bin/go run main.go -nodetype 2 & # Consensus
   /usr/local/go/bin/go run main.go -nodetype 3  & # Replica
done
# go run main.go -nodetype 0  -numberProps 1 & # Leader
# go run main.go -nodetype 1 -nodeId 0  &# Proposer
# go run main.go -nodetype 2 & # Consensus
# go run main.go -nodetype 2  & # Consensus
# go run main.go -nodetype 2  &# Consensus
# go run main.go -nodetype 3  &

echo "I HATE THIS"
# Flood leader with requests


# Get performance throughput


# Scale up to:
# 1 leader, 100 proposers, 50 consensus nodes, 50 replicas


# Flood leader with same number of requests


# Get performance throughput

