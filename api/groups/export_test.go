package groups

import chainApiShared "github.com/multiversx/mx-chain-go/api/shared"

// HandleErrorAndReturn -
func HandleHTTPError(err string) (int, chainApiShared.ReturnCode) {
	return handleHTTPError(err)
}
