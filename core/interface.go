package core

import (
	"github.com/ElrondNetwork/elrond-sdk-erdgo/data"
)

type Guardian interface {
	ValidateAndSend(transaction data.Transaction) (string, error)
}
