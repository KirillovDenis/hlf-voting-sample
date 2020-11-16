#!/bin/bash

export PATH=$GOPATH/fabric-samples/bin:$PATH
export FABRIC_CFG_PATH=${PWD}
CHANNEL_NAME=mychannel


mv ./crypto-config/IssuerPublicKey ./crypto-config/peerOrganizations/idemix-config/msp/IssuerPublicKey
mv ./crypto-config/RevocationPublicKey ./crypto-config/peerOrganizations/idemix-config/msp/RevocationPublicKey



configtxgen -profile SoloOrgsOrdererGenesis -outputBlock ./config/genesis.block -channelID sys-channel
if [ "$?" -ne 0 ]; then
  echo "Failed to generate orderer genesis block..."
  exit 1
fi

configtxgen -profile SoloOrgsChannel -outputCreateChannelTx ./config/channel.tx -channelID $CHANNEL_NAME
if [ "$?" -ne 0 ]; then
  echo "Failed to generate channel configuration transaction..."
  exit 1
fi

configtxgen -profile SoloOrgsChannel -outputAnchorPeersUpdate ./config/Org1MSPanchors.tx -channelID $CHANNEL_NAME -asOrg Org1MSP
if [ "$?" -ne 0 ]; then
  echo "Failed to generate anchor peer update for Org1MSP..."
  exit 1
fi

# configtxgen -profile SoloOrgsChannel -outputAnchorPeersUpdate ./config/Org2MSPanchors.tx -channelID $CHANNEL_NAME -asOrg Org1MSP
# if [ "$?" -ne 0 ]; then
#   echo "Failed to generate anchor peer update for Org2MSP..."
#   exit 1
# fi
