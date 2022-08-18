package providers

import "github.com/ElrondNetwork/elrond-sdk-erdgo/data"

type Codes map[string]string

type GuardianValidateRequest struct {
	Account string `json:"account"`
	Codes   struct {
		Totp string `json:"totp"`
	} `json:"codes"`
	Tx data.Transaction `json:"transaction"`
}

type GuardianRegisterRequest struct {
	Account string `json:"account"`
	Type    string `json:"type"`
}
