package main

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric/core/chaincode/shim/ext/cid"
)

func (t *voting) getComposite(stub shim.ChaincodeStubInterface) (string, error) {
	return stub.CreateCompositeKey(constVoting, []string{t.ID})
}

func createVoting(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("Incorrect number of arguments. Expecting 1 (voting)")
	}

	voting := &voting{}
	if err := json.Unmarshal([]byte(args[0]), voting); err != nil {
		return nil, err
	}

	// voting.ID = uuid.New().String()

	privKey, err := getPrivateKey(2048)
	if err != nil {
		return nil, err
	}
	encodedSKStr, err := encodePrivateKeyStr(privKey)
	if err != nil {
		return nil, err
	}
	encodedPKStr, err := encodePublicKeyStr(&privKey.PublicKey)
	if err != nil {
		return nil, err
	}
	compositeKey, err := getCompositeRegisterKey(stub, voting.ID)
	if err != nil {
		return nil, err
	}

	if err := stub.PutPrivateData(privateKeysConst, compositeKey, []byte(*encodedSKStr)); err != nil {
		return nil, err
	}

	voting.PubKey = *encodedPKStr

	if err := saveState(stub, voting); err != nil {
		return nil, err
	}

	return []byte(voting.ID), nil
}

func getBlindSign(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("Incorrect number of arguments. Expecting 1 (UserRegData)")
	}

	regData := &userRegData{}
	err := json.Unmarshal([]byte(args[0]), regData)
	if err != nil {
		return nil, err
	}

	cert, err := cid.GetX509Certificate(stub)
	if err != nil {
		return nil, err
	}

	if cert.Subject.CommonName != regData.UserID {
		return nil, fmt.Errorf("Your names are not matched")
	}

	voting := &voting{ID: regData.VotingID}
	if err := loadState(stub, voting); err != nil {
		return nil, err
	}

	if !contains(voting.Participants, regData.UserID) {
		return nil, fmt.Errorf("Your are not participant in this voting")
	}

	compositeSignedUser, err := getCompositeSignedUser(stub, voting.ID, regData.UserID)
	if err != nil {
		return nil, err
	}

	if existState(stub, compositeSignedUser) {
		return nil, fmt.Errorf("You already signed for this voting")
	}

	if err := stub.PutState(compositeSignedUser, []byte{0x00}); err != nil {
		return nil, err
	}

	compositeRegisterKey, err := getCompositeRegisterKey(stub, voting.ID)
	if err != nil {
		return nil, err
	}
	privateKeyRaw, err := stub.GetPrivateData(privateKeysConst, compositeRegisterKey)
	if err != nil {
		return nil, err
	}
	privateKeyStr := string(privateKeyRaw)

	privateKey, err := decodePrivateKeyStr(privateKeyStr)
	if err != nil {
		return nil, err
	}

	signedMsg := encrypt(regData.Data, privateKey.D, privateKey.N)

	return signedMsg, nil
}

func registerUserInVotingIdemix(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("Incorrect number of arguments. Expecting 1 (userTallingData)")
	}

	regTallingData := &userTallingData{}
	err := json.Unmarshal([]byte(args[0]), regTallingData)
	if err != nil {
		return nil, err
	}

	voting := &voting{ID: regTallingData.VotingID}
	if err := loadState(stub, voting); err != nil {
		return nil, err
	}

	votingPubKey, err := decodePublicKeyStr(voting.PubKey)
	if err != nil {
		return nil, err
	}
	userPubKey, err := decodePublicKeyStr(regTallingData.Key)
	if err != nil {
		return nil, err
	}

	if err := checkSign(userPubKey, regTallingData.SignedHashKey, votingPubKey); err != nil {
		return nil, err
	}

	compositeRegisteredUser, err := getCompositeRegisteredUserIdemix(stub, regTallingData.VotingID, regTallingData.Key)
	if err != nil {
		return nil, err
	}

	if existState(stub, compositeRegisteredUser) {
		return nil, fmt.Errorf("You have already registered your key")
	}

	if err := putStateInLedger(stub, compositeRegisteredUser, []byte{0x00}); err != nil {
		return nil, err
	}

	return nil, nil
}

func voteIdemix(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("Incorrect number of arguments. Expecting 1 (votingBallot)")
	}

	ballot := &votingBallot{}
	if err := json.Unmarshal([]byte(args[0]), ballot); err != nil {
		return nil, err
	}

	ou, found, err := cid.GetAttributeValue(stub, "ou")
	if err != nil {
		return nil, fmt.Errorf("Failed to get attribute 'ou'")
	}
	if !found {
		return nil, fmt.Errorf("attribute 'ou' not found")
	}
	logger.Infof("Organizational unit: '%s'", ou)

	pubKey, err := decodePublicKeyStr(ballot.Key)
	if err != nil {
		return nil, err
	}

	if err := checkSign(pubKey, ballot.SignedKey, pubKey); err != nil {
		return nil, err
	}

	compositeRegisteredUser, err := getCompositeRegisteredUserIdemix(stub, ballot.VotingID, ballot.Key)
	if err != nil {
		return nil, err
	}

	if !existState(stub, compositeRegisteredUser) {
		return nil, fmt.Errorf("You are not registered in voting '%s'", ballot.VotingID)
	}

	voteCompositeKey, err := getCompositeVote(stub, ballot.VotingID, ballot.Key)
	if err != nil {
		return nil, err
	}

	if existState(stub, voteCompositeKey) {
		return nil, fmt.Errorf("You already voted")
	}

	if err := putStateInLedger(stub, voteCompositeKey, ballot); err != nil {
		return nil, err
	}

	return nil, nil
}

func getResults(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("Incorrect number of arguments. Expecting 1 (votingId)")
	}

	voting := &voting{ID: args[0]}
	if err := loadState(stub, voting); err != nil {
		return nil, err
	}

	iterator, err := stub.GetStateByPartialCompositeKey(voting.ID, []string{constBallots})
	if err != nil {
		return nil, err
	}
	defer iterator.Close()

	// Initializing results
	results := &votingResults{Questions: make(map[string]map[string][]string)}
	for _, question := range voting.Questions {
		results.Questions[question.Title] = make(map[string][]string)
		for _, answer := range question.Answers {
			results.Questions[question.Title][answer.Value] = []string{}
		}
	}

	// Counting results
	for iterator.HasNext() {
		queryResultKV, err := iterator.Next()
		if err != nil {
			return nil, err
		}

		_, arr, err := stub.SplitCompositeKey(queryResultKV.Key)
		if err != nil {
			return nil, err
		}
		userKey := arr[len(arr)-1]
		logger.Infof("ballot user '%s' ", userKey)

		currentBallot := votingBallot{}
		err = json.Unmarshal(queryResultKV.GetValue(), &currentBallot)
		if err != nil {
			return nil, err
		}

		for _, question := range currentBallot.Questions {
			results.updateResults(question, userKey)
		}

	}

	return marshal(results)
}

func (results *votingResults) updateResults(question votingQuestion, userKey string) {
	_, ok := results.Questions[question.Title]
	if !ok {
		results.Questions[question.Title] = make(map[string][]string)
	}

	for _, answer := range question.Answers {
		_, ok := results.Questions[question.Title][answer.Value]
		if !ok {
			results.Questions[question.Title][answer.Value] = []string{userKey}
		} else {
			results.Questions[question.Title][answer.Value] = append(results.Questions[question.Title][answer.Value], userKey)
		}
	}
}
