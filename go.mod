module github.com/ElrondNetwork/multi-factor-auth-go-service

go 1.15

require (
	github.com/ElrondNetwork/elrond-go v1.3.36
	github.com/ElrondNetwork/elrond-go-core v1.1.15
	github.com/ElrondNetwork/elrond-go-logger v1.0.7
	github.com/btcsuite/websocket v0.0.0-20150119174127-31079b680792
	github.com/gin-contrib/cors v0.0.0-20190301062745-f9e10995c85a
	github.com/gin-contrib/pprof v1.3.0
	github.com/gin-gonic/gin v1.8.0
	github.com/sec51/convert v1.0.2 // indirect
	github.com/sec51/cryptoengine v0.0.0-20180911112225-2306d105a49e // indirect
	github.com/sec51/gf256 v0.0.0-20160126143050-2454accbeb9e // indirect
	github.com/sec51/qrcode v0.0.0-20160126144534-b7779abbcaf1 // indirect
	github.com/sec51/twofactor v1.0.0
	github.com/stretchr/testify v1.7.1
	github.com/urfave/cli v1.22.9
	google.golang.org/protobuf v1.28.1 // indirect
)

replace github.com/ElrondNetwork/arwen-wasm-vm/v1_2 v1.2.40 => github.com/ElrondNetwork/arwen-wasm-vm v1.2.40

replace github.com/ElrondNetwork/arwen-wasm-vm/v1_3 v1.3.40 => github.com/ElrondNetwork/arwen-wasm-vm v1.3.40

replace github.com/ElrondNetwork/arwen-wasm-vm/v1_4 v1.4.54-rc3 => github.com/ElrondNetwork/arwen-wasm-vm v1.4.54-rc3
