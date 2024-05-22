#!/bin/bash

set -x

payload="$(dd if=/dev/urandom bs=3 count=1)"

mainnet="https://tools.multiversx.com"
devnet="https://devnet-tools.multiversx.com"
testnet="https://testnet-tools.multiversx.com"
upmad="https://up-mad-auth.elrond.ro"
localhost="http://localhost:8080"
host=${testnet}

curl -X 'POST' \
  "${host}/guardian/sign-transaction" \
  -H 'accept: application/json' \
  -H 'Authorization: Bearer ZXJkMWU5NDA1MG1hc3VoeXY3aGQzeDhqbGc4eHJ6ZGttM2N0czhqaGF3ODRoa2ZqcndneDRhcHNhZnUyZWM=.ZEdWemRHbHVady45MGE2Y2ZiY2IxMTRiN2RkOTU2M2Y5ZTYzYmJmNDYyMzVlNDI3ODUwMjVjNWJhNzJjMzIzZjM2NmJiNzdkMmY2Ljg2NDAwLmUzMA.743dc79fc66112769447a148bd5a403336bf82f7e29d640fbdd8cf3d5e12637256992155d7138c3e04f3ff17924042a5496e6497ca51072eca204ceaff469b06' \
  -H 'Content-Type: application/json' \
  -d '{
  "code": "323532",
  "second-code": "514956",
  "transaction": {
    "chainID": "T",
    "gasLimit": 500000,
    "gasPrice": 1000000,
    "guardian": "erd1x5tndhj5q0jlftj4e58gt638xwuckhtzrgy5ktr2gfd0zhth0f9qwvktcf",
    "guardianSignature": "",
    "nonce": 0,
    "options": 0,
    "data": "",
    "receiver": "erd1e94050masuhyv7hd3x8jlg8xrzdkm3cts8jhaw84hkfjrwgx4apsafu2ec",
    "sender": "erd1e94050masuhyv7hd3x8jlg8xrzdkm3cts8jhaw84hkfjrwgx4apsafu2ec",
    "signature": "bff7bf5a0e75f918644785c00c9b7de2852c3a89f75587cd46f8e843c35e0f6ff3a3ccf23567283d20fc51e0f8583553c44bd9cf849e223e9216b9464617650b",
    "value": "0",
    "version": 2,
    "options": 2
  }
}' -vvv
