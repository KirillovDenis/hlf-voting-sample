FROM golang:1.11.5

ARG FAB_VERSION=1.4.1
ARG COUCHDB_VERSION=0.4.15

ENV FABRIC_ROOT=$GOPATH/src/github.com/hyperledger/fabric
ENV FABRIC_CFG_PATH=/etc/hyperledger/fabric

RUN mkdir -p $GOPATH/src/github.com/hyperledger/fabric \
         $FABRIC_CFG_PATH

RUN apt-get update \
        && curl -sSL http://bit.ly/2ysbOFE | bash -s -- ${FAB_VERSION} ${FAB_VERSION} ${COUCHDB_VERSION} \
        && apt-get install -y vim

RUN cd $GOPATH/src/github.com/hyperledger \
        && git clone --single-branch -b release-1.4 --depth 1 http://gerrit.hyperledger.org/fabric 

RUN go get -u github.com/google/uuid


ENV FABRIC_ROOT=$GOPATH/src/github.com/hyperledger/fabric 
WORKDIR $FABRIC_ROOT/../fabric-network
CMD ["bash"]

