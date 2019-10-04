echo "##########################################################"
echo "#####        building fabric host container      #########"
echo "##########################################################"
docker image build -t "dltc/fabrichost_test:latest" .

docker run -dit --name fabric_host_test.com \
  -v $PWD:/go/src/github.com/hyperledger/fabric-network/ \
  dltc/fabrichost_test

mkdir -p $PWD/fabric/core/chaincode/shim/ext/cid 
docker cp fabric_host_test.com:/go/src/github.com/hyperledger/fabric/core/chaincode/shim/ext/cid $PWD/fabric/core/chaincode/shim/ext/cid 
mkdir -p $PWD/fabric/core/chaincode/lib/cid
docker cp fabric_host_test.com:/go/src/github.com/hyperledger/fabric/core/chaincode/lib/cid $PWD/fabric/core/chaincode/lib/cid
mkdir -p $PWD/fabric/google/
docker cp fabric_host_test.com:/go/src/github.com/google/uuid $PWD/fabric/google/uuid


mkdir config && mkdir crypto-config


echo "##########################################################"
echo "#####           generate crypto meterials        #########"
echo "##########################################################"
docker exec fabric_host_test.com bash -c ./scripts/generate1.sh

echo "##########################################################"
echo "#####       starting ca and copy idemix keys     #########"
echo "##########################################################"

docker-compose -f compose-files/solo.yaml up -d ca.test.com
sleep 3

docker cp ca.test.com:/etc/hyperledger/fabric-ca-server/IssuerPublicKey ${PWD}/crypto-config/IssuerPublicKey
docker cp ca.test.com:/etc/hyperledger/fabric-ca-server/IssuerRevocationPublicKey ${PWD}/crypto-config/RevocationPublicKey

echo "##########################################################"
echo "#####           generate channel artifacts       #########"
echo "##########################################################"
docker exec fabric_host_test.com ./scripts/generate2.sh 



docker-compose -f compose-files/solo.yaml up -d
sleep 3



docker exec cli.test.com ./scripts/createChannel.sh 


# peer chaincode invoke -C mychannel -n votingcc -c '{"Args":["initVoting"]}' -o orderer.test.com:7050

