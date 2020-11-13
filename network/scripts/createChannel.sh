#!/bin/bash

CHANNEL_NAME=mychannel
CC_SRC_PATH=github.com/chaincode/
VER_CC=1.0
CC_NAME=votingcc
ACT=instantiate
ORDERER_PEER=orderer1.sample.com:7050
    
echo "##########################################################"
echo "#####              Create channel                #########"
echo "##########################################################"
peer channel create -o $ORDERER_PEER -c $CHANNEL_NAME -f /opt/gopath/src/github.com/hyperledger/fabric/peer/configtx/channel.tx

   
echo "##########################################################"
echo "#####              Join channel 0 1              #########"
echo "##########################################################" 
peer channel join -b $CHANNEL_NAME.block 

echo "##########################################################"
echo "#####            Updating channel 1              #########"
echo "##########################################################"
peer channel update -o $ORDERER_PEER -c $CHANNEL_NAME -f /opt/gopath/src/github.com/hyperledger/fabric/peer/configtx/Org1MSPanchors.tx

echo "##########################################################"
echo "#####        Installing chaincode  1             #########"
echo "##########################################################"    
peer chaincode install -n "$CC_NAME" -v "$VER_CC" -p "$CC_SRC_PATH"

echo "##########################################################"
echo "#####        Instantiating chaincode 1           #########"
echo "##########################################################"
peer chaincode "$ACT" -o $ORDERER_PEER -C $CHANNEL_NAME -n "$CC_NAME"  -v "$VER_CC" -c '{"Args":["init"]}' -P "OR ('Org1MSP.member')" \
 --collections-config /opt/gopath/src/github.com/chaincode/collection_config.json
