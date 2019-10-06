package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/hyperledger/fabric/core/chaincode/shim"
)

func getStruct(stub shim.ChaincodeStubInterface, key string, obj interface{}) ([]byte, error) {
	logger.Infof("Getting state by key: '%s'", key)

	objRaw, err := stub.GetState(key)
	if objRaw == nil && err == nil {
		return nil, fmt.Errorf("Key: %s doesn't exist", key)
	}

	if err := json.Unmarshal(objRaw, obj); err != nil {
		return nil, err
	}

	return objRaw, nil
}

func encodePrivateKey(key *rsa.PrivateKey) ([]byte, error) {
	bytes, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		return nil, err
	}

	return bytes, nil
}

func encodePrivateKeyStr(key *rsa.PrivateKey) (*string, error) {
	bytes, err := encodePrivateKey(key)
	if err != nil {
		return nil, nil
	}
	encoded := base64.StdEncoding.EncodeToString(bytes)
	return &encoded, nil
}

func encodePublicKey(key *rsa.PublicKey) ([]byte, error) {
	bytes, err := x509.MarshalPKIXPublicKey(key)
	if err != nil {
		return nil, err
	}

	return bytes, nil
}

func encodePublicKeyStr(key *rsa.PublicKey) (*string, error) {
	bytes, err := encodePublicKey(key)
	if err != nil {
		return nil, err
	}
	encoded := base64.StdEncoding.EncodeToString(bytes)
	return &encoded, nil
}

func getPrivateKey(bits int) (*rsa.PrivateKey, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return nil, err
	}
	return privateKey, nil
}

func decodePrivateKey(pvtData []byte) (*rsa.PrivateKey, error) {
	key, err := x509.ParsePKCS8PrivateKey(pvtData)
	if err != nil {
		return nil, err
	}

	rsaKey, ok := key.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("Can't decode private rsa key")
	}

	return rsaKey, nil
}

func decodePrivateKeyStr(encodedStr string) (*rsa.PrivateKey, error) {
	bytes, err := base64.StdEncoding.DecodeString(encodedStr)
	if err != nil {
		return nil, err
	}

	return decodePrivateKey(bytes)
}

func decodePublicKey(pubData []byte) (*rsa.PublicKey, error) {
	key, err := x509.ParsePKIXPublicKey(pubData)
	if err != nil {
		return nil, err
	}

	rsaKey, ok := key.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("Can't decode public rsa key")
	}
	return rsaKey, nil
}

func decodePublicKeyStr(encodedStr string) (*rsa.PublicKey, error) {
	bytes, err := base64.StdEncoding.DecodeString(encodedStr)
	if err != nil {
		return nil, err
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

func existState(stub shim.ChaincodeStubInterface, key string) bool {
	raw, err := stub.GetState(key)
	if raw == nil && err == nil {
		return false
	}
	return true
}

func getCompositeSignedUser(stub shim.ChaincodeStubInterface, votingID string, userID string) (string, error) {
	return stub.CreateCompositeKey(votingID, []string{constSigneddUser, userID})
}

func getCompositeRegisteredUserIdemix(stub shim.ChaincodeStubInterface, votingID string, userID string) (string, error) {
	return stub.CreateCompositeKey(votingID, []string{constRegisteredUser, userID})
}

func getCompositeRegisterKey(stub shim.ChaincodeStubInterface, localVotingID string) (string, error) {
	return stub.CreateCompositeKey(registerKeyConst, []string{localVotingID})
}

func getCompositeVote(stub shim.ChaincodeStubInterface, votingID string, key string) (string, error) {
	return stub.CreateCompositeKey(votingID, []string{constBallots, key})
}

func marshal(i interface{}) ([]byte, error) {
	raw, err := json.Marshal(i)
	if err != nil {
		return nil, err
	}

	return raw, nil
}

func loadState(stub shim.ChaincodeStubInterface, model keyModel) error {
	composite, err := model.getComposite(stub)
	if err != nil {
		return err
	}

	if _, err := getStruct(stub, composite, model); err != nil {
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

func checkSign(userPubKey *rsa.PublicKey, signedHashKey string, votingPubKey *rsa.PublicKey) error {
	logger.Info("checking signature")
	logger.Info("user public key:")
	logger.Info("module: ", userPubKey.N.String())
	logger.Info("exp: ", userPubKey.E)

	logger.Info("voting public key:")
	logger.Info("module: ", votingPubKey.N.String())
	logger.Info("exp: ", votingPubKey.E)

	logger.Info("signed hash: ", signedHashKey)

	E := big.NewInt(int64(votingPubKey.E))
	decryptedHash := encryptBigInt(signedHashKey, E, votingPubKey.N)

	logger.Info("decrypted hash: ", decryptedHash.String())

	hexHash, err := getHash(userPubKey)
	if err != nil {
		return err
	}
	logger.Info("hex hash: ", hexHash)
	bytes, err := hex.DecodeString(*hexHash)
	if err != nil {
		return err
	}

	hashedKey := &big.Int{}
	hashedKey.SetBytes(bytes)

	logger.Info("hex hash big int: ", hashedKey.String())

	if decryptedHash.Cmp(hashedKey) == 0 {
		return nil
	}
	return fmt.Errorf("Invalid signature")
}

func getHash(key *rsa.PublicKey) (*string, error) {
	bytesForHash, err := encodePublicKey(key)
	if err != nil {
		return nil, err
	}

	result := fmt.Sprintf("%x", sha256.Sum256(bytesForHash))
	return &result, nil
}

func encrypt(message string, exp *big.Int, module *big.Int) []byte {
	result, msg := &big.Int{}, &big.Int{}
	msg.SetString(message, 10)

	result.Exp(msg, exp, module)
	logger.Info("signed data: ", result.String())

	return []byte(result.String())
}

func encryptBigInt(message string, exp *big.Int, module *big.Int) *big.Int {
	result, msg := &big.Int{}, &big.Int{}
	msg.SetString(message, 10)

	result.Exp(msg, exp, module)
	logger.Info("signed data: ", result.String())

	return result
}
