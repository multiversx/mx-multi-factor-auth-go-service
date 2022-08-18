package providers

type Codes map[string]string

type GuardianValidateRequest struct {
	Account string `json:"account"`
	Codes   struct {
		Totp string `json:"totp"`
	} `json:"codes"`
}

type GuardianRegisterRequest struct {
	Account string `json:"account"`
	Type    string `json:"type"`
}
