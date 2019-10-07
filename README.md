# hlf-voting-sample
Samples for Hyperledger Bootcamp session "HLF Identity Mixer in secret e-voting"

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
Run the following command
```
cd client/hlf-voting-sample

mvn exec:java
```

### Cleaning
In order to remove network change working directory to "network" and run the following script
```
./clear.sh
```



