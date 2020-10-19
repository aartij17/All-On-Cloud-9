![Caper logo](https://github.com/AartiJivrajani/All-On-Cloud-9/blob/master/images/logo.png?raw=true)  
Supply chain management using Cross Application Permissioned Blockchain. The project is an attempt to implement [this work](https://dl.acm.org/doi/pdf/10.14778/3342263.3342275).

# Consideration
The project will assume presence of 4 applications -
1. Manufacturer
2. Supplier
3. Bulk Buyer
4. Carrier
---------------------------------------------------------------------------------------------------------------------
BLOCKCHAIN LEDGER
---------------------------------------------------------------------------------------------------------------------
The Caper Blockchain is a SINGLE blockchain which is maintained across all the applications. BUT, each application
can view only the LOCAL view of the blockchain which consists of -
- Internal transactions of the application
- Cross application transactions

The Blockchain ledger maintains 3 properties:
- Total order between all internal transactions
- Total order between all external transactions
- Internal transaction might include a cryptographic hash of an external transaction on which it depends(as described above).

The blockchain ledger is modeled in the form of a DAG where the first block is the genesis block.
If t comes AFTER t', t' ----> t [transaction t includes H(t')] where H(.) is the cryptographic hash

TIP: After every consensus run, when the transaction is executed on each node of the applications, an update has to be
made which updates the view of the blockchain for that particular blockchain.

---------------------------------------------------------------------------------------------------------------------
BLOCK
---------------------------------------------------------------------------------------------------------------------
- Type:
	- internal transaction
	- external transaction
- Edge
	- Incoming edges (to blocks)
	- Outgoing edges (to blocks)
- Transaction number/ID
	- Internal transaction of an application: t11, t12, t13 are internal transactions of application A1. 
	- Cross application transactions: t<12, 1>, t<23,2> are cross-application transactions of applications A1 and A2. 
	- Each cross application transaction is labeled with t<i, j> where i indicates order of transaction among transactions
	  that are initiated by the initiator application and j represents the order of the transaction among all
	  cross-application transactions. [CHECK FIGURE 2 IN THE PAPER FOR MORE DETAILS]

---------------------------------------------------------------------------------------------------------------------
TRANSACTION
---------------------------------------------------------------------------------------------------------------------
- Consists of the cryptographic hash of the previous transaction
- To exhibit data dependency(internal transaction depending on other cross application transaction), the
  cryptographic hash of the external transaction is included in the internal transaction(Example: manufacturer calculates
  materials demand based on the place order transaction of Bulk Buyer).

---------------------------------------------------------------------------------------------------------------------
APPLICATION
---------------------------------------------------------------------------------------------------------------------
An application maintains the following:
- Datastore
- View of the blockchain ledger
- Private Smart Contracts
	- to implement the application business logic
- Public Smart Contracts
	- To implement logic of the cross application transactions
	- These smart contracts run on EVERY application. i.e., they are generic/global.
	- They include several conditions which need to be checked before the transactions are executed at each of the nodes
	(after the consensus protocol runs).
	- They include actions for non-conforming transactions where the initiator application is penalized if the SLA
	is not adhered to.

---------------------------------------------------------------------------------------------------------------------
DEPLOYMENT
---------------------------------------------------------------------------------------------------------------------
- Each application has (3f+1) nodes(aka agents). So in this case, each application should have atleast 4 nodes.
- None of these nodes should overlap. i.e., for any two applications a and b, Na(set of a's agents) interesect Nb is NULL.
- The nodes communicate which each other using bidirectional secure channels.
- To ensure safety and liveness for cross application transactions, at least 2/3rd of the applications INCLUDING the
initiator application must agree on the order of the transaction - i.e.: (2/3A + 1) applications.

---------------------------------------------------------------------------------------------------------------------
MESSAGES
---------------------------------------------------------------------------------------------------------------------
- Contain public key signatures
- Contain the message digest: cryptographic hash of the message(using a good hash algorithm).
- <m>r : represents message m signed by replica r
- D(m): Message digest of message m
- Assume all nodes have public keys of all other machines

