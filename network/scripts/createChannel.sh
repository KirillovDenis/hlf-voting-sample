#!/bin/bash

CHANNEL_NAME=mychannel
CC_SRC_PATH=github.com/chaincode/
VER_CC=1.0
CC_NAME=votingcc
ACT=instantiate
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

joinOrgsToChannel() {
	for org in 1 2; do
        echo "##########################################################"
        echo "#####              Join channel   $org              #########"
        echo "##########################################################"        
		setOrg $org
        peer channel join -b mychannel.block 
	done
}



echo "##########################################################"
echo "#####              Create channel                #########"
echo "##########################################################"
setOrg 1
peer channel create -o $ORDERER_PEER -c $CHANNEL_NAME -f /opt/gopath/src/github.com/hyperledger/fabric/peer/configtx/channel.tx

   
joinOrgsToChannel

echo "##########################################################"
echo "#####            Updating channel 1              #########"
echo "##########################################################"
setOrg 1
peer channel update -o $ORDERER_PEER -c $CHANNEL_NAME -f /opt/gopath/src/github.com/hyperledger/fabric/peer/configtx/Org1MSPanchors.tx

echo "##########################################################"
echo "#####            Updating channel 2              #########"
echo "##########################################################"
setOrg 2
peer channel update -o $ORDERER_PEER -c $CHANNEL_NAME -f /opt/gopath/src/github.com/hyperledger/fabric/peer/configtx/Org2MSPanchors.tx



echo "##########################################################"
echo "#####        Installing chaincode  1             #########"
echo "##########################################################"
setOrg 1    
peer chaincode install -n "$CC_NAME" -v "$VER_CC" -p "$CC_SRC_PATH"


echo "##########################################################"
echo "#####        Installing chaincode  2             #########"
echo "##########################################################"    
setOrg 2
peer chaincode install -n "$CC_NAME" -v "$VER_CC" -p "$CC_SRC_PATH"



echo "##########################################################"
echo "#####        Instantiating chaincode 1           #########"
echo "##########################################################"
setOrg 1
peer chaincode "$ACT" -o $ORDERER_PEER -C $CHANNEL_NAME -n "$CC_NAME"  -v "$VER_CC" -c '{"Args":["init"]}' -P "OR ('Org1MSP.member', 'idemixMSPID1.member)" \
 --collections-config /opt/gopath/src/github.com/chaincode/collection_config.json
