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

### DESIGN

Each application that is deployed has 2 types of transactions -  
- Internal transactions: These can be modeled as investments, fiat entities put on hold etc.  
- Cross-application transactions: These can be simple transactions like lending/borrowing money  