#### SAMPLE CODE ONLY. NOT FOR DEPLOYMENT

This is minimalistic code which uses nats as a message broker. 

In order to ensure failure handling, please take a look at the nats.Options struct

Open 3 terminals and run the following commands in each
1. Start the nats server  
Install nats server if you dont have it
```bash
brew install nats-server
```  
Start the server  
```nats-server```

2. Start the server
```
cd server
go run main.go
```

3. Start the client
```
cd client
go run main.go
```