#!/bin/bash

VER_CC=1.1
CHANNEL_NAME=mychannel
CC_SRC_PATH=github.com/chaincode/
CC_NAME=votingcc
ACT=upgrade
ORDERER_PEER=orderer1.sample.com:7050
PEER0_ORG1_CA=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.sample.com/peers/peer0.org1.sample.com/tls/ca.crt
PEER0_ORG2_CA=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org2.sample.com/peers/peer0.org2.sample.com/tls/ca.crt

    

setOrg() {
    ORG=$1
    if [ $ORG -eq 1 ]; then
        CORE_PEER_LOCALMSPID="Org1MSP"
        CORE_PEER_TLS_ROOTCERT_FILE=$PEER0_ORG1_CA
        CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.sample.com/users/Admin@org1.sample.com/msp
        CORE_PEER_ADDRESS=peer0.org1.sample.com:7051
    elif [ $ORG -eq 2 ]; then
        CORE_PEER_LOCALMSPID="Org2MSP"
        CORE_PEER_TLS_ROOTCERT_FILE=$PEER0_ORG2_CA
        CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org2.sample.com/users/Admin@org2.sample.com/msp
        CORE_PEER_ADDRESS=peer0.org2.sample.com:9051
    fi
}

echo "Upgrading chaincode to version $VER_CC"

docker exec -e 'CORE_PEER_LOCALMSPID=Org1MSP' \
-e 'CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.sample.com/users/Admin@org1.sample.com/msp' \
-e 'CORE_PEER_ADDRESS=peer0.org1.sample.com:7051' \
cli.sample.com peer chaincode install -n "$CC_NAME" -v "$VER_CC" -p "$CC_SRC_PATH" 
  
# docker exec -e 'CORE_PEER_LOCALMSPID=Org2MSP' \
# -e 'CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org2.sample.com/users/Admin@org2.sample.com/msp' \
# -e 'CORE_PEER_ADDRESS=peer0.org2.sample.com:9051' \
# cli.sample.com peer chaincode install -n "$CC_NAME" -v "$VER_CC" -p "$CC_SRC_PATH" 


docker exec cli.sample.com peer chaincode $ACT -o $ORDERER_PEER -C $CHANNEL_NAME -n "$CC_NAME"  -v "$VER_CC" -c '{"Args":["init"]}' -P "OR ('Org1MSP.member', 'idemixMSPID1.member')" \
    --collections-config /opt/gopath/src/github.com/chaincode/collection_config.json