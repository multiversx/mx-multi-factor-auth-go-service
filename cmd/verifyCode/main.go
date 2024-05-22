package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base32"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/multiversx/mx-multi-factor-auth-go-service/core"
	"github.com/multiversx/mx-multi-factor-auth-go-service/core/requests"
	"github.com/urfave/cli"
)

const (
	tcsAddress         = "https://tcs-api-ams.elrond.ro"
	verifyCodeEndpoint = "guardian/verify-code"

	httpUserAgentKey = "User-Agent"
	httpUserAgent    = "MultiversX testing"

	httpAcceptTypeKey = "Accept"
	httpAcceptType    = "application/json"

	httpContentTypeKey = "Content-Type"
	httpContentType    = "application/json"

	httpAuthorizationKey = "Authorization"
	httpBearerToken      = "Bearer ZXJkMXlrcWQ2NGZ4eHBwNHdzejB2N3NqcWVtMDM4d2Zwemxsamh4NG1od3g4dzlsY3htZHpjZnN6cnA2NGE.ZEdWemRHbHVady5hMDc3YTdiNDBiNDZhYzE4M2M1NmY5MGFkYmVlYWJmMjQwZTAwMmYzNmE0MjM1NmNjMDIyNjVjODQ2NzNhZDZhLjg2NDAwLmUzMA.2ab768909fc5b157a50de44dbff0a5a37f8b758113f4c8c4dd358f29c91d5a3b1b936ab4478b7087c66b93384f1f2137bafbd2bc44b10553b7c03a13268baa04"

	httpRealIPKey = "X-Real-IP"

	guardian = "erd1whx9e025h44u6czjhmnrq40796dsyfzvw3wd9aj83m0yqec78xtsujfsdc"

	otpDigits = 6
	otpPeriod = 30
	otpSecret = "C2QVNVYXNCUJDM3CFPQ67LIBMQ4KEMKS"
)

var (
	client = http.DefaultClient

	goodCodes = cli.IntFlag{
		Name:  "good-codes",
		Usage: "Int option of how many codes to be correct. Options: 0, 1 or 2",
		Value: 0,
	}
)

type verifyCodeFailureResponse struct {
	Data  requests.OTPCodeVerifyDataResponse `json:"data"`
	Error string                             `json:"error"`
	Code  string                             `json:"code"`
}

type verifyCodeOkResponse struct {
	Data  string `json:"data"`
	Error string `json:"error"`
	Code  string `json:"code"`
}

func main() {
	app := cli.NewApp()
	app.Name = "Verify code CLI app"
	app.Flags = []cli.Flag{
		goodCodes,
	}
	app.Authors = []cli.Author{
		{
			Name:  "The MultiversX Team",
			Email: "contact@multiversx.com",
		},
	}

	app.Action = func(c *cli.Context) error {
		return verifyCodeRequest(c)
	}

	err := app.Run(os.Args)
	if err != nil {
		println(err.Error())
		os.Exit(1)
	}
}

func verifyCodeRequest(c *cli.Context) error {
	numberOfGoodCodes := c.GlobalInt(goodCodes.Name)

	codes, err := generateCodes(numberOfGoodCodes)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s/%s", tcsAddress, verifyCodeEndpoint)

	payload := requests.VerificationPayload{
		Code:       codes[0],
		SecondCode: codes[1],
		Guardian:   guardian,
	}
	payloadBuff, err := json.Marshal(&payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewReader(payloadBuff))
	if err != nil {
		return err
	}

	applyPostHeaderParams(req)
	applyBearerParam(req)
	applyRandomIPParam(req)

	doResp, err := client.Do(req)
	if err != nil {
		return err
	}

	buff, err := io.ReadAll(doResp.Body)
	if err != nil {
		return err
	}

	if len(buff) == 0 {
		return fmt.Errorf("%w while calling %s, code %d", core.ErrEmptyData, url, doResp.StatusCode)
	}

	if doResp.StatusCode == http.StatusTooManyRequests ||
		doResp.StatusCode == http.StatusBadRequest {
		var resp verifyCodeFailureResponse
		err = json.Unmarshal(buff, &resp)
		if err != nil {
			return err
		}

		data, err := json.Marshal(resp.Data)
		if err != nil {
			return err
		}

		println(fmt.Sprintf("%s, %s", resp.Error, string(data)))

		return nil
	}

	var resp verifyCodeOkResponse
	err = json.Unmarshal(buff, &resp)
	if err != nil {
		return err
	}

	data, err := json.Marshal(resp)
	if err != nil {
		return err
	}

	println(fmt.Sprintf("Code ok, %s", string(data)))

	return nil
}

func applyGetHeaderParams(request *http.Request) {
	request.Header.Set(httpAcceptTypeKey, httpAcceptType)
	request.Header.Set(httpUserAgentKey, httpUserAgent)
}

func applyPostHeaderParams(request *http.Request) {
	applyGetHeaderParams(request)
	request.Header.Set(httpContentTypeKey, httpContentType)
}

func applyBearerParam(request *http.Request) {
	request.Header.Set(httpAuthorizationKey, httpBearerToken)
}

func applyRandomIPParam(request *http.Request) {
	request.Header.Set(httpRealIPKey, randIP())
}

func randIP() string {
	return fmt.Sprintf("0.0.0.%d", rand.Intn(200))
}

func generateCodes(goodCodes int) ([]string, error) {
	codes, err := generateGoodCodes(goodCodes)
	if err != nil {
		return nil, err
	}

	randomCodesNeeded := 2 - len(codes)

	for i := 0; i < randomCodesNeeded; i++ {
		codes = append(codes, generateRandomCode())
	}

	return codes, nil
}

func generateRandomCode() string {
	code := ""
	for i := 0; i < otpDigits; i++ {
		code += fmt.Sprintf("%d", rand.Intn(9))
	}
	return code
}

func generateGoodCodes(numOfCodes int) ([]string, error) {
	if numOfCodes == 0 {
		return []string{}, nil
	}

	key, err := base32.StdEncoding.DecodeString(otpSecret)
	if err != nil {
		return nil, err
	}

	codes := make([]string, 0, numOfCodes)
	counter := time.Now().Unix() / otpPeriod
	for i := 0; i < numOfCodes; i++ {
		code, err := generateCode(counter, key)
		if err != nil {
			return nil, err
		}

		codes = append(codes, code)
		counter++
	}

	return codes, nil
}

func generateCode(counter int64, key []byte) (p string, err error) {
	hash := hmac.New(sha1.New, key)

	err = binary.Write(hash, binary.BigEndian, counter)
	if err != nil {
		return "", err
	}

	h := hash.Sum(nil)
	offset := h[19] & 0x0f
	trunc := binary.BigEndian.Uint32(h[offset : offset+4])
	trunc &= 0x7fffffff
	code := trunc % uint32(math.Pow(10, float64(otpDigits)))

	return fmt.Sprintf("%0"+strconv.Itoa(otpDigits)+"d", code), nil
}
