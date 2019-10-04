package main

import (
	"fmt"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric/protos/peer"
)

var logger *shim.ChaincodeLogger

// VotingChaincode common chaincode for initializing voting process
type VotingChaincode struct {
}

var bcFunctions = map[string]func(shim.ChaincodeStubInterface, []string) peer.Response{
	"createVoting": createVoting,
}

// Init chaincode interface
// ========================================
func (t *VotingChaincode) Init(stub shim.ChaincodeStubInterface) peer.Response {
	logger = shim.NewLogger("loger")
	return shim.Success(nil)
}

// Invoke - Our entry point for Invocations
// ========================================
func (t *VotingChaincode) Invoke(stub shim.ChaincodeStubInterface) peer.Response {
	function, args := stub.GetFunctionAndParameters()
	if logger == nil {
		logger = shim.NewLogger("loger")
	}

	logger.Infof("=================Start function '%s'=================", function)
	logger.Infof("Args: '%s'", args)
	logger.Info("--------------------------------------------------------------------")

	bcFunc := bcFunctions[function]
	if bcFunc == nil {
		logger.Error("Invoke did not find func: " + function)
		return shim.Error("Received unknown function invocation.")
	}

	response := bcFunc(stub, args)
	logger.Info("=================End function=================")
	return response
}

// Just main stub
// ===============
func main() {
	err := shim.Start(new(VotingChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}
