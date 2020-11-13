#!/bin/bash

VER_CC=1.1
CHANNEL_NAME=mychannel
CC_SRC_PATH=github.com/chaincode/
CC_NAME=votingcc
ACT=upgrade
ORDERER_PEER=orderer1.sample.com:7050

echo "Upgrading chaincode to version $VER_CC"
docker exec cli.sample.com peer chaincode install -n "$CC_NAME" -v "$VER_CC" -p "$CC_SRC_PATH" 
  
docker exec cli.sample.com peer chaincode $ACT -o $ORDERER_PEER -C $CHANNEL_NAME -n "$CC_NAME"  -v "$VER_CC" -c '{"Args":["init"]}' -P "OR ('Org1MSP.member', 'idemixMSPID1.member')" \
    --collections-config /opt/gopath/src/github.com/chaincode/collection_config.json