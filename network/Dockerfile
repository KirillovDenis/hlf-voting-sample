FROM golang:1.11.5

ENV FABRIC_ROOT=$GOPATH/src/github.com/hyperledger/fabric
ENV FABRIC_CFG_PATH=/etc/hyperledger/fabric

RUN mkdir -p $GOPATH/src/github.com/hyperledger/fabric \
         $FABRIC_CFG_PATH

RUN apt-get update \
        && curl -sSL http://bit.ly/2ysbOFE | bash -s -- 1.4.1 1.4.1 0.4.15 \
        && apt-get install -y vim

RUN cd $GOPATH/src/github.com/hyperledger \
        && git clone --single-branch -b release-1.4 --depth 1 http://gerrit.hyperledger.org/fabric 

RUN go get -u github.com/google/uuid


ENV FABRIC_ROOT=$GOPATH/src/github.com/hyperledger/fabric 
WORKDIR $FABRIC_ROOT/../fabric-network
CMD ["bash"]

