package main

import (
	"encoding/json"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric/core/chaincode/shim/ext/cid"
	"github.com/hyperledger/fabric/protos/peer"
)

func (t *voting) getComposite(stub shim.ChaincodeStubInterface) (string, error) {
	return stub.CreateCompositeKey(constVoting, []string{t.ID})
}

func createVoting(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1 (voting)")
	}

	voting := &voting{}
	if err := json.Unmarshal([]byte(args[0]), voting); err != nil {
		logger.Error(err)
		return shim.Error(err.Error())
	}

	// voting.ID = uuid.New().String()

	if err := saveState(stub, voting); err != nil {
		logger.Error(err)
		return shim.Error(err.Error())
	}

	return shim.Success([]byte(voting.ID))
}

func getBlindSign(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1 (UserRegData)")
	}

	regData := &userRegData{}
	err := json.Unmarshal([]byte(args[0]), regData)
	if err != nil {
		logger.Error(err)
		return shim.Error(err.Error())
	}

	cert, err := cid.GetX509Certificate(stub)
	if err != nil {
		logger.Error(err)
		return shim.Error(err.Error())
	}

	if cert.Subject.CommonName != regData.UserID {
		return shim.Error("Your names are not matched")
	}

	voting := &voting{ID: regData.VotingID}
	if err := loadState(stub, voting); err != nil {
		logger.Error(err)
		return shim.Error(err.Error())
	}

	if !contains(voting.Participants, regData.UserID) {
		return shim.Error("Your are not participant in this voting")
	}

	compositeSignedUser, err := getCompositeSignedUser(stub, voting.ID, regData.UserID)
	if err != nil {
		logger.Error(err)
		return shim.Error(err.Error())
	}

	if existState(stub, compositeSignedUser) {
		return shim.Error("You already signed for this voting")
	}

	if err := stub.PutState(compositeSignedUser, []byte{0x00}); err != nil {
		logger.Error(err)
		return shim.Error(err.Error())
	}

	compositeRegisterKey, err := getCompositeRegisterKey(stub, voting.ID)
	if err != nil {
		logger.Error(err)
		return shim.Error(err.Error())
	}
	privateKeyRaw, err := stub.GetPrivateData(privateKeysConst, compositeRegisterKey)
	if err != nil {
		logger.Error(err)
		return shim.Error(err.Error())
	}
	privateKeyStr := string(privateKeyRaw)

	privateKey, err := decodePrivateKeyStr(privateKeyStr)
	if err != nil {
		logger.Error(err)
		return shim.Error(err.Error())
	}

	signedMsg := encrypt(regData.Data, privateKey.D, privateKey.N)

	return shim.Success(signedMsg)
}
