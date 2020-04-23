## NATS for inter-node communication

- Request-reply messaging pattern suits the requirement of this project quite well [https://docs.nats.io/developing-with-nats/tutorials/reqreply]


#### Deploying NATS server  
##### Pre-requisites: Make sure you have docker up and running on your system

Pull the docker image from dockerhub
```bash
docker pull nats
```
Run the docker container
```bash
docker images | grep nats
Select the container ID
# map the internal ports to make them accessible from outside the docker environment
docker run -p 4222:4222 -p 8222:8222 <container_ID> &  
```

## DESIGN

### Application
Each application will have 3 nodes deployed with IDs of the format: ```NODE<app_id><node #>```. For example ```NODE12```
denotes ```node 2``` of ```application 1```.  
```Node 1``` of each application will act as the ```PRIMARY NODE``` - this node receives requests from clients.  
Each application that is deployed has 2 types of transactions -  
- Internal transactions: These can be modeled as investments, fiat entities put on hold etc.  
- Cross-application transactions: These can be simple transactions like lending/borrowing money  

### Blockchain
The blockchain is modeled as a Directed Acyclic Graph  
Each block maintains a single transaction  
Each application can view 2 types of blockchain -  
- Public: All applications have the same view of this blockchain. This blockchain can only be updated by ```CROSS_APPLICATION_TXN```
- Private: Maintained separately per app. An application can both R/W to this blockchain.  

### Transactions
Two types of transactions:
- ```CROSS_APPLICATION_TXN```  
- ```INTERNAL_TXN```

#### QUESTIONS
1. Is there any difference between public/private records and public/private blockchains? 