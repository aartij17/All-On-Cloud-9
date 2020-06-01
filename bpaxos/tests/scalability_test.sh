#!/bin/bash

rm -f examine.txt
rm -f statuslog.txt
# Base case:
/usr/local/go/bin/go run main.go -nodetype 0  -numberProps 5 > examine.txt & # Leader

for j in `seq 0 4`;
do
    /usr/local/go/bin/go run main.go -nodetype 1 -nodeId $j & # Proposer
done

for i in `seq 0 49`;
do
   /usr/local/go/bin/go run main.go -nodetype 2 & # Consensus
   /usr/local/go/bin/go run main.go -nodetype 3  & # Replica
done
nsus

echo "I HATE THIS"

