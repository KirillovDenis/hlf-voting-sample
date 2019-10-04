package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"strings"

	"github.com/hyperledger/fabric/core/chaincode/shim/ext/cid"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric/protos/peer"
)

func getStructNew(stub shim.ChaincodeStubInterface, key string, obj interface{}) ([]byte, error) {
	var errorText string

	fmt.Println("getting by key ", key)

	objRaw, err := stub.GetState(key)
	if objRaw == nil && err == nil {
		errorText = fmt.Sprintf("Key: %s doesn't exist.", key)
		logger.Error(errorText)
		return nil, errors.New(errorText)
	}

	err = json.Unmarshal(objRaw, obj)
	if err != nil {
		errorText = fmt.Sprintf("Error while unmarshaling obj: %s. %s", key, err.Error())
		logger.Error(errorText)
		return nil, errors.New(errorText)
	}

	return objRaw, nil
}

func test(stub shim.ChaincodeStubInterface, args []string) peer.Response {

	return shim.Success([]byte(args[0]))
}

func encodePrivateKey(key *rsa.PrivateKey) []byte {
	bytes, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		logger.Error(err.Error())
		panic(err)
	}

	return bytes
}

func encodePrivateKeyStr(key *rsa.PrivateKey) string {
	bytes := encodePrivateKey(key)
	encoded := base64.StdEncoding.EncodeToString(bytes)
	return encoded
}

func encodePublicKey(key *rsa.PublicKey) []byte {
	bytes, err := x509.MarshalPKIXPublicKey(key)
	if err != nil {
		logger.Error(err.Error())
		panic(err)
	}

	return bytes
}

func encodePublicKeyStr(key *rsa.PublicKey) string {
	bytes := encodePublicKey(key)
	encoded := base64.StdEncoding.EncodeToString(bytes)
	return encoded
}

func getPrivateKey(bits int) *rsa.PrivateKey {
	privateKey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		fmt.Println(err)
	}
	return privateKey
}

func decodePrivateKey(pvtData []byte) (*rsa.PrivateKey, error) {
	key, err := x509.ParsePKCS8PrivateKey(pvtData)
	if err != nil {
		logger.Error(err.Error())
		return nil, err
	}

	rsaKey, ok := key.(*rsa.PrivateKey)
	if !ok {
		msg := "Can't decode private rsa key"
		logger.Error(msg)
		return nil, err
	}

	return rsaKey, nil
}

func decodePrivateKeyStr(encodedStr string) (*rsa.PrivateKey, error) {
	bytes, err := base64.StdEncoding.DecodeString(encodedStr)
	if err != nil {
		logger.Error(err.Error())
		return nil, err
	}

	return decodePrivateKey(bytes)
}

func decodePublicKey(pubData []byte) *rsa.PublicKey {
	key, err := x509.ParsePKIXPublicKey(pubData)
	if err != nil {
		logger.Error(err.Error())
		panic(err)
	}

	rsaKey, ok := key.(*rsa.PublicKey)
	if !ok {
		msg := "Can't decode public rsa key"
		logger.Error(msg)
		panic(msg)
	}
	return rsaKey
}

func decodePublicKeyStr(encodedStr string) *rsa.PublicKey {
	bytes, err := base64.StdEncoding.DecodeString(encodedStr)
	if err != nil {
		logger.Error(err.Error())
		panic(err)
	}

	return decodePublicKey(bytes)
}

// Array of string containce function
func contains(slice []string, element string) bool {
	for _, el := range slice {
		if el == element {
			return true
		}
	}
	return false
}

func containsPtr(slice []*string, element *string) bool {
	for _, el := range slice {
		if el == element {
			return true
		}
	}
	return false
}

func containsAnswer(slice *[]votingAnswer, ansValue string) bool {
	for _, el := range *slice {
		if el.Value == ansValue {
			return true
		}
	}
	return false
}

func containsWithIndex(slice []string, element string) (bool, int) {
	for i, el := range slice {
		if el == element {
			return true, i
		}
	}
	return false, -1
}

func existState(stub shim.ChaincodeStubInterface, key string) bool {
	raw, err := stub.GetState(key)
	if raw == nil && err == nil {
		return false
	}
	return true
}

func decryptBallot(msg string, rsaKey *rsa.PrivateKey) (string, error) {
	bound := 344
	var rerults strings.Builder

	if len(msg) > bound {
		for i := 0; i < len(msg)/bound; i++ {
			currentSubString := msg[i*bound : (i+1)*bound]
			decryptedRaw, err := decryptBase64(currentSubString, rsaKey)
			if err != nil {
				return "", err
			}

			rerults.Write(decryptedRaw)
		}
		if (len(msg)/bound)*bound != len(msg) {
			currentSubString := msg[(len(msg)/bound)*bound:]
			decryptedRaw, err := decryptBase64(currentSubString, rsaKey)
			if err != nil {
				return "", err
			}

			rerults.Write(decryptedRaw)
		}

	} else {
		decryptedRaw, err := decryptBase64(msg, rsaKey)
		if err != nil {
			return "", err
		}

		rerults.Write(decryptedRaw)
	}

	return rerults.String(), nil
}

func decryptBase64(msg string, rsaKey *rsa.PrivateKey) ([]byte, error) {
	msgRaw, err := base64.StdEncoding.DecodeString(msg)
	if err != nil {
		return nil, err
	}

	decryptedRaw, err := rsa.DecryptPKCS1v15(rand.Reader, rsaKey, msgRaw)
	if err != nil {
		return nil, err
	}

	return decryptedRaw, nil
}

func getPrivateKeyForDepartment(stub shim.ChaincodeStubInterface, localVotingID string) (*rsa.PrivateKey, error) {
	compositeRegisterKey, err := getCompositeRegisterKey(stub, localVotingID)
	if err != nil {
		return nil, err
	}
	privateKeyRaw, err := stub.GetPrivateData(privateKeysConst, compositeRegisterKey)
	if err != nil {
		return nil, err
	}
	privateKeyStr := string(privateKeyRaw)

	fmt.Println("private key: ", privateKeyStr)

	return decodePrivateKeyStr(privateKeyStr)
}

func getCompositeSharing(stub shim.ChaincodeStubInterface, localVotingID string) (string, error) {
	org, err := getOrgFromStub(stub)
	if err != nil {
		return "", err
	}

	return stub.CreateCompositeKey(org, []string{constSharingkey, localVotingID})
}

func getCompositePublicSharing(stub shim.ChaincodeStubInterface, localVotingID string) (string, error) {
	org, err := getOrgFromStub(stub)
	if err != nil {
		return "", err
	}

	return stub.CreateCompositeKey(org, []string{constPublicSharingKey, localVotingID})
}

func getCompositeSignedUser(stub shim.ChaincodeStubInterface, localVotingID string, userName string) (string, error) {
	org, err := getOrgFromStub(stub)
	if err != nil {
		return "", err
	}

	return stub.CreateCompositeKey(org, []string{localVotingID, constSigneddUser, userName})
}

func getCompositeRegisteredUser(stub shim.ChaincodeStubInterface, localVotingID string) (string, error) {
	org, userName, err := getOrgNameFromStub(stub)
	if err != nil {
		return "", err
	}

	return stub.CreateCompositeKey(org, []string{localVotingID, constRegisteredUser, userName})
}

func getCompositeRegisteredUserPoll(stub shim.ChaincodeStubInterface, pollID string, userID string) (string, error) {
	return stub.CreateCompositeKey(constRegisteredUser, []string{pollID, userID})
}

func getCompositeRegisteredUserIdemix(stub shim.ChaincodeStubInterface, localVotingID string, userName string) (string, error) {
	org, err := getOrgFromStub(stub)
	if err != nil {
		return "", err
	}

	return stub.CreateCompositeKey(org, []string{localVotingID, constRegisteredUser, userName})
}

func getCompositeRegisteredUsers(stub shim.ChaincodeStubInterface, localVotingID string) (string, error) {
	org, err := getOrgFromStub(stub)
	if err != nil {
		return "", err
	}

	return stub.CreateCompositeKey(org, []string{registeredUsersConst, localVotingID})
}

func getCompositeRegisterKey(stub shim.ChaincodeStubInterface, localVotingID string) (string, error) {
	org, err := getOrgFromStub(stub)
	if err != nil {
		return "", err
	}

	return stub.CreateCompositeKey(org, []string{registerKeyConst, localVotingID})
}

func getCompositeRegisterPubKey(stub shim.ChaincodeStubInterface, localVotingID string) (string, error) {
	org, err := getOrgFromStub(stub)
	if err != nil {
		return "", err
	}

	return stub.CreateCompositeKey(org, []string{registerKeyPubConst, localVotingID})
}

func getCompositeResultsKey(stub shim.ChaincodeStubInterface, localVotingID string) (string, error) {
	org, err := getOrgFromStub(stub)
	if err != nil {
		return "", err
	}

	return stub.CreateCompositeKey(org, []string{resultsKeyConst, localVotingID})
}

func getCompositeVoting(stub shim.ChaincodeStubInterface, votingID string) (string, error) {
	org, err := getOrgFromStub(stub)
	if err != nil {
		return "", err
	}

	return stub.CreateCompositeKey(org, []string{constVoting, votingID})
}

func getCompositeLocal(stub shim.ChaincodeStubInterface, localVotingID string) (string, error) {
	org, err := getOrgFromStub(stub)
	if err != nil {
		return "", err
	}

	return stub.CreateCompositeKey(org, []string{constLocalVoting, localVotingID})
}

func getCompositeLocalStatistics(stub shim.ChaincodeStubInterface, localVotingID string) (string, error) {
	org, err := getOrgFromStub(stub)
	if err != nil {
		return "", err
	}

	return stub.CreateCompositeKey(org, []string{constLocalVotingStatistics, localVotingID})
}

func getCompositeLocalByID(stub shim.ChaincodeStubInterface, id string) (string, error) {
	org, err := getOrgFromStub(stub)
	if err != nil {
		return "", err
	}

	return stub.CreateCompositeKey(org, []string{constLocalVoting, id})
}

func getCompositeObservers(stub shim.ChaincodeStubInterface, votingID string) (string, error) {
	org, err := getOrgFromStub(stub)
	if err != nil {
		return "", err
	}

	return stub.CreateCompositeKey(org, []string{constObservers, votingID})
}

func getCompositeVote(stub shim.ChaincodeStubInterface, localVotingID string, key string) (string, error) {
	org, err := getOrgFromStub(stub)
	if err != nil {
		return "", err
	}

	return stub.CreateCompositeKey(org, []string{localVotingID, constBallots, key})
}

func getCompositeVotePoll(stub shim.ChaincodeStubInterface, pollID string, userID string) (string, error) {
	return stub.CreateCompositeKey(constBallots, []string{pollID, userID})
}

func getCompositeUsersList(stub shim.ChaincodeStubInterface) (string, error) {
	return constUsersList, nil
}

func getCompositeUserNotify(stub shim.ChaincodeStubInterface, userID string, pollID string) (string, error) {
	return stub.CreateCompositeKey(constNotification, []string{userID, pollID})
}

func getCompositeGroupsList(stub shim.ChaincodeStubInterface) (string, error) {
	return constGroupsList, nil
}

func getCompositeObserversList(stub shim.ChaincodeStubInterface) (string, error) {
	return constObserversList, nil
}

func getCompositeUserName(stub shim.ChaincodeStubInterface, userName string) (string, error) {
	org, err := getOrgFromStub(stub)
	if err != nil {
		return "", err
	}

	return stub.CreateCompositeKey(org, []string{userName})
}

func getOrgFromStub(stub shim.ChaincodeStubInterface) (string, error) {
	cert, err := cid.GetX509Certificate(stub)
	if err != nil {
		return "", err
	}

	if cert != nil {
		if len(cert.Subject.OrganizationalUnit) > 1 {
			return cert.Subject.OrganizationalUnit[1], nil
		}

		return "", fmt.Errorf("Not found any organizational unit")
	}

	ou, found, err := cid.GetAttributeValue(stub, "ou")
	if err != nil {
		msg := "Failed to get attribute 'ou'"
		logger.Error(msg)
		return "", fmt.Errorf(msg)
	}

	if !found {
		msg := "attribute 'ou' not found"
		return "", fmt.Errorf(msg)
	}

	return ou, nil
}

func getOrgNameFromStub(stub shim.ChaincodeStubInterface) (string, string, error) {
	cert, err := cid.GetX509Certificate(stub)
	if err != nil {
		return "", "", err
	}

	if cert != nil {
		if len(cert.Subject.OrganizationalUnit) > 1 {
			return cert.Subject.OrganizationalUnit[1], cert.Subject.CommonName, nil
		}

		return "", "", fmt.Errorf("Not found any organizational unit")
	}

	ou, found, err := cid.GetAttributeValue(stub, "ou")
	if err != nil {
		msg := "Failed to get attribute 'ou'"
		logger.Error(msg)
		return "", "", fmt.Errorf(msg)
	}

	if !found {
		msg := "attribute 'ou' not found"
		return "", "", fmt.Errorf(msg)
	}

	return ou, cert.Subject.CommonName, nil
}

func getBytes(i interface{}) peer.Response {
	raw, err := json.Marshal(i)
	if err != nil {
		logger.Error(err)
		return shim.Error(err.Error())
	}

	return shim.Success(raw)
}

func getNameFromStub(stub shim.ChaincodeStubInterface) (string, error) {
	cert, err := cid.GetX509Certificate(stub)
	if err != nil {
		return "", err
	}

	return cert.Subject.CommonName, nil
}

func peerResponse(i interface{}) peer.Response {
	logger.Info("Invoked 'peerResponse' function")

	iRaw, err := json.Marshal(i)
	if err != nil {
		logger.Error(err)
		return shim.Error(err.Error())
	}

	return shim.Success(iRaw)
}

func loadState(stub shim.ChaincodeStubInterface, model keyModel) error {
	composite, err := model.getComposite(stub)
	if err != nil {
		return err
	}

	if _, err := getStructNew(stub, composite, model); err != nil {
		return err
	}

	return nil
}

func saveState(stub shim.ChaincodeStubInterface, model keyModel) error {
	compositeKey, err := model.getComposite(stub)
	if err != nil {
		return err
	}

	if err := putStateInLedger(stub, compositeKey, model); err != nil {
		return err
	}

	return nil
}

func putStateInLedger(stub shim.ChaincodeStubInterface, key string, i interface{}) error {
	objRaw, err := json.Marshal(i)
	if err != nil {
		return err
	}

	logger.Infof("State to write. Key: '%s', Value: '%s'", key, string(objRaw))

	err = stub.PutState(key, objRaw)
	if err != nil {
		return err
	}

	return nil
}

func checkSign(userPubKey *rsa.PublicKey, signedHashKey string, votingPubKey *rsa.PublicKey) bool {
	fmt.Println("checking signature")
	fmt.Println("user public key:")
	fmt.Println("module: ", userPubKey.N.String())
	fmt.Println("exp: ", userPubKey.E)

	fmt.Println("voting public key:")
	fmt.Println("module: ", votingPubKey.N.String())
	fmt.Println("exp: ", votingPubKey.E)

	fmt.Println("signed hash: ", signedHashKey)

	E := big.NewInt(int64(votingPubKey.E))
	decryptedHash := encryptBigInt(signedHashKey, E, votingPubKey.N)

	fmt.Println("decrypted hash: ", decryptedHash.String())

	hexHash := getHash(userPubKey)
	fmt.Println("hex hash: ", hexHash)
	bytes, err := hex.DecodeString(hexHash)
	if err != nil {
		logger.Error(err.Error())
	}

	hashedKey := &big.Int{}
	hashedKey.SetBytes(bytes)

	fmt.Println("hex hash big int: ", hashedKey.String())

	return decryptedHash.Cmp(hashedKey) == 0
}

func signRevokedKey(key *rsa.PublicKey, singerKey *rsa.PrivateKey) string {
	bigHashKey := getHashBigStr(key)
	bigHashKey.Exp(bigHashKey, singerKey.D, singerKey.N)

	return bigHashKey.String()
}

func getHash(key *rsa.PublicKey) string {
	bytesForHash := encodePublicKey(key)
	result := sha256.Sum256(bytesForHash)

	return fmt.Sprintf("%x", result)
}

func getHashBigStr(key *rsa.PublicKey) *big.Int {
	hexHash := getHash(key)
	bytes, err := hex.DecodeString(hexHash)
	if err != nil {
		logger.Error(err.Error())
	}

	hashedKey := &big.Int{}
	hashedKey.SetBytes(bytes)
	return hashedKey
}

func getHashSHA256(data string) string {
	result := sha256.Sum256([]byte(data))
	return fmt.Sprintf("%x", result)
}

func checkRegisteredKeyExisting(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2 (localVotingId, key)")
	}

	localVotingID := args[0]
	key := args[1]

	ou, found, err := cid.GetAttributeValue(stub, "ou")
	if err != nil {
		msg := "Failed to get attribute 'ou'"
		logger.Error(msg)
		shim.Error(msg)
	}
	if !found {
		msg := "attribute 'ou' not found"
		logger.Error(msg)
		return shim.Error(err.Error())
	}

	compositeRegisteredUser, err := stub.CreateCompositeKey(ou, []string{localVotingID, constRegisteredUser, key})
	if err != nil {
		logger.Error(err)
		return shim.Error(err.Error())
	}

	if existState(stub, compositeRegisteredUser) {
		return shim.Success([]byte("true"))
	}

	return shim.Success([]byte("false"))
}

func encrypt(message string, exp *big.Int, module *big.Int) []byte {
	result, msg := &big.Int{}, &big.Int{}
	msg.SetString(message, 10)

	result.Exp(msg, exp, module)
	fmt.Println("signed data: ", result.String())

	return []byte(result.String())
}

func encryptBigInt(message string, exp *big.Int, module *big.Int) *big.Int {
	result, msg := &big.Int{}, &big.Int{}
	msg.SetString(message, 10)

	result.Exp(msg, exp, module)
	fmt.Println("signed data: ", result.String())

	return result
}
