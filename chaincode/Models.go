package main

import (
	"time"

	"github.com/hyperledger/fabric/core/chaincode/shim"
)

type voting struct {
	Title        string           `json:"title"`
	ID           string           `json:"id"`
	Questions    []votingQuestion `json:"questions"`
	Participants []string         `json:"participants"`
	Results      *votingResults   `json:"results"`
	Time         votingTime       `json:"time"`
}

type votingQuestion struct {
	Title   string         `json:"title"`
	Answers []votingAnswer `json:"answers"`
}

type votingAnswer struct {
	Value string `json:"value"`
}

type votingTime struct {
	VotingStart       time.Time `json:"votingStart"`
	VotingEnd         time.Time `json:"votingEnd"`
	RegistrationStart time.Time `json:"registrationStart"`
	RegistrationEnd   time.Time `json:"registrationEnd"`
}

type votingResults struct {
	Questions    map[string]map[string][]string `json:"questions"`
	YetEncrypted bool                           `json:"yetEncrypted"`
}

type votingBallot struct {
	UserID    string           `json:"userId"`
	VotingID  string           `json:"votingId"`
	Questions []votingQuestion `json:"questions"`
}

type keyModel interface {
	getComposite(shim.ChaincodeStubInterface) (string, error)
}

type userRegData struct {
	UserID   string `json:"userId"`
	VotingID string `json:"votingId"`
	Data     string `json:"data"`
}
