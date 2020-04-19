# All-On-Cloud-9
CS293B: Cloud Computing, Edge Computing and IoT Project 

### Solidity files for the smart contracts

Smart contracts are written in Solidity

- In order to use the contract in the Go source code, we must first generate the ABI(Application Binary Interface)
of the contract and compile the ABI to a format which can be imported in our Go code. 

- For this, download the Solidity compiler `solc`    
```bash
MacOS: 
brew update
brew tap ethereum/ethereum
brew install solidity

Ubuntu:
sudo snap install solc --edge
```
- We also need to install a tool called Abigen for generating the ABI from Solidity smart contract  
```bash
go get -u github.com/ethereum/go-ethereum
cd $GOPATH/src/github.com/ethereum/go-ethereum/
make
make devtools
```

- Once the smart contract is created, generate the ABI from a solidity source file. It will write to a file - `./build/Store.abi`  
```bash
solc --abi Store.sol -o build
```  

- Convert the ABI to a Go file that can be imported. This new file will contain all the available methods that we can  
use to interact with the smart contract from our Go application
```bash
abigen --abi=./build/Store.abi --pkg=store --out=Store.go
```

- In order to deploy a smart contract from Go, we also need to compile the solidity smart contract to EVM bytecode.  
The EVM bytecode is what will be sent in the data field of the transaction. The bin file is required for generating the  
deploy methods on the Go contract file.
```bash
solc --bin Store.sol -o build
```

- Now we compile the Go contract file which will include the deploy methods because we includes the bin file.  
```bash
abigen --bin=./build/Store.bin --abi=./build/Store.abi --pkg=store --out=Store.go
```

