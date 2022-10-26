module github.com/ElrondNetwork/multi-factor-auth-go-service

go 1.15

require (
	github.com/ElrondNetwork/elrond-go v1.3.38-0.20220912122303-9c2574322163
	github.com/ElrondNetwork/elrond-go-core v1.1.20-0.20220912122639-a040477c8cb7
	github.com/ElrondNetwork/elrond-go-crypto v1.0.1
	github.com/ElrondNetwork/elrond-go-logger v1.0.7
	github.com/ElrondNetwork/elrond-sdk-erdgo v1.0.24-0.20220927113814-d155226b0bf6
	github.com/btcsuite/websocket v0.0.0-20150119174127-31079b680792
	github.com/gin-contrib/cors v0.0.0-20190301062745-f9e10995c85a
	github.com/gin-contrib/pprof v1.3.0
	github.com/gin-gonic/gin v1.8.0
	github.com/go-errors/errors v1.0.1
	github.com/sec51/convert v1.0.2 // indirect
	github.com/sec51/cryptoengine v0.0.0-20180911112225-2306d105a49e // indirect
	github.com/sec51/gf256 v0.0.0-20160126143050-2454accbeb9e // indirect
	github.com/sec51/qrcode v0.0.0-20160126144534-b7779abbcaf1 // indirect
	github.com/sec51/twofactor v1.0.0
	github.com/stretchr/testify v1.7.1
	github.com/urfave/cli v1.22.9
	google.golang.org/protobuf v1.28.1 // indirect
)

replace github.com/ElrondNetwork/arwen-wasm-vm/v1_2 v1.2.41 => github.com/ElrondNetwork/arwen-wasm-vm v1.2.42-0.20220825092831-7d45c37a8a73

replace github.com/ElrondNetwork/arwen-wasm-vm/v1_3 v1.3.41 => github.com/ElrondNetwork/arwen-wasm-vm v1.3.42-0.20220825091352-272f48a2c23c

replace github.com/ElrondNetwork/arwen-wasm-vm/v1_4 v1.4.58 => github.com/ElrondNetwork/arwen-wasm-vm v1.4.59-0.20220825090722-70fbc73c9021
