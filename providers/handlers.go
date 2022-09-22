package providers

import (
	"crypto"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/sec51/twofactor"
)

func readOtps(filename string) (map[string][]byte, error) {
	data, err := os.ReadFile(fmt.Sprintf("%s.json", filename))
	if err != nil {
		return nil, err
	}
	var otpsEncoded map[string][]byte
	err = json.Unmarshal(data, &otpsEncoded)

	return otpsEncoded, err
}

func saveOtp(filename string, otps map[string][]byte) error {
	filePath := fmt.Sprintf("%s.json", filename)
	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE, 0777)
	defer closeFile(file)
	if err != nil {
		log.Println(err)
		return err
	}
	jsonOtps, err := json.Marshal(otps)

	_, err = file.Write(jsonOtps)
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func closeFile(file *os.File) {
	err := file.Close()
	if err != nil {
		log.Println(err)
	}
}

func newTOTP(account, issuer string, hash crypto.Hash, digits int) (Totp, error) {
	return twofactor.NewTOTP(account, issuer, hash, digits)
}

func totpFromBytes(encryptedMessage []byte, issuer string) (Totp, error) {
	return twofactor.TOTPFromBytes(encryptedMessage, issuer)
}
