#!/bin/bash

export PATH=$GOPATH/fabric-samples/bin:$PATH
export FABRIC_CFG_PATH=${PWD}
CHANNEL_NAME=mychannel

# generate crypto material
cryptogen generate --config=./crypto-config.yaml
if [ "$?" -ne 0 ]; then
  echo "Failed to generate crypto material..."
  exit 1
fi

mkdir -p crypto-config/peerOrganizations/idemix-config/msp/

CA_KEY=$(ls ./crypto-config/peerOrganizations/org1.test.com/ca/ | grep -i sk)
echo $CA_KEY

mv ./crypto-config/peerOrganizations/org1.test.com/ca/$CA_KEY  ./crypto-config/peerOrganizations/org1.test.com/ca/ca-key.pem
mv ./crypto-config/peerOrganizations/org1.test.com/ca/ca.org1.test.com-cert.pem  ./crypto-config/peerOrganizations/org1.test.com/ca/ca-cert.pem

