package guardian

import "github.com/ElrondNetwork/multi-factor-auth-go-service/config"

func CreateMockGuardianConfig() *config.GuardianConfig {
	return &config.GuardianConfig{
		PrivateKeyFile:       "testdata/grace.pem",
		RequestTimeInSeconds: 2,
	}
}
