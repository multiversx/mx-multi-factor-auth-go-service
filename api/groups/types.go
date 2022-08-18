package groups

// GeneralMetrics represents an objects metrics map
type GeneralMetrics map[string]interface{}

type Codes map[string]string

type GuardianValidateRequest struct {
	Account string `json:"account"`
	Codes   Codes  `json:"codes"`
}

type GuardianRegisterRequest struct {
	Account string `json:"account"`
	Type    string `json:"type"`
}
