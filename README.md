# hlf-voting-sample
Simple Hyperledger Fabric application.
There are two organization:

1. org1 (can send proposals and invoke chaincode)
2. org2 (read-only, can view blocks)

## Getting Started

### Installing
In order to raise sample network:

1. Copy this repository and change working directory to "network"
```
git clone git@github.com:KirillovDenis/hlf-voting-sample.git

cd hlf-voting-sample/network
```

2. Up network
```
./raiseNetwork.sh
```

### Invoke chaincode
To interact with chaincode java sdk are used.
Run application.
```
cd client/hlf-voting-sample

mvn exec:java
```

### Read blocks
To read block need change 'org1' client organization to 'org2' in 'client/hlf-voting-samples/src/main/resources/network-config.yaml'
Then run
#### 1. Update client organization
Go to the resource folder
```
cd client/hlf-voting-sample/src/main/resources
```
Change 'org1' client organization to 'org2' in 'network-config.yaml' file

#### 2. Run application
```
cd client/hlf-voting-sample

mvn exec:java
```

### Cleaning
In order to remove network change working directory to "network" and run the following script
```
./clear.sh
```



