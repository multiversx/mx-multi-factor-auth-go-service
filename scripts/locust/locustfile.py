import base64

from locust import task, between
from locust.contrib.fasthttp import FastHttpUser
from multiversx_sdk_wallet import Mnemonic
from multiversx_sdk_core import Address, Transaction

import os
import json
from pathlib import Path

import pyotp
import requests


def generateBech32AddressAsBase64():
    mnemonic = Mnemonic.generate()
    secret_key = mnemonic.derive_key(0)
    public_key = secret_key.generate_public_key()
    raw_public_key = bytes.fromhex(public_key.hex())

    address = Address(raw_public_key, "erd")

    b = base64.b64encode(bytes(address.bech32(), 'utf-8'))

    return b.decode("utf-8")


def generateMockAuthorizationToken():
    address = generateBech32AddressAsBase64()
    b = base64.b64encode(bytes("erd1e94050masuhyv7hd3x8jlg8xrzdkm3cts8jhaw84hkfjrwgx4apsafu2ec", 'utf-8'))
    address = b.decode("utf-8")

    auth_token = "Bearer " + address + ".ZEdWemRHbHVady45MGE2Y2ZiY2IxMTRiN2RkOTU2M2Y5ZTYzYmJmNDYyMzVlNDI3ODUwMjVjNWJhNzJjMzIzZjM2NmJiNzdkMmY2Ljg2NDAwLmUzMA.743dc79fc66112769447a148bd5a403336bf82f7e29d640fbdd8cf3d5e12637256992155d7138c3e04f3ff17924042a5496e6497ca51072eca204ceaff469b06"

    header = {'Authorization': auth_token, 'Content-Type': 'application/json'}
    return header


def read_wallet_files(path):
    wallets = []

    dir_list = os.listdir(path)
    for file_name in dir_list:
        with open(path+file_name) as json_file:
            address = file_name.split(".")[0]
            wallet_data = json.load(json_file)
            wallet_data["address"] = address
            wallets.append(wallet_data)

    return wallets


def create_tx(address, guardian):
    tx = Transaction(
        sender=address,
        receiver=address,
        guardian=guardian,
        value=0,
        gas_limit=500000,
        gas_price=10000000,
        chain_id="T",
        version=2,
        options=2
    )

    return tx.to_dictionary()


def generate_otp_code(secret):
    totp = pyotp.TOTP(secret)
    return totp.now()


class WebsiteUser(FastHttpUser):
    wait_time = between(1, 2)

    @task(1)
    def register_mock_auth(self):
        header = generateMockAuthorizationToken()

        print(header)
        resp = self.client.post("/guardian/register", headers=header, json={'tag':'darius'})
        print(resp.content)
        print(resp.request.headers)
        print(resp.request.url)

    def sign_tx_from_wallets(self):
        headers = {
            'accept': 'application/json',
            'Content-Type': 'application/json',
        }

        path = "../wallets/"
        wallets = read_wallet_files(path)

        for wallet in wallets:
            data = {}

            address = Address.from_bech32(wallet["address"])
            guardian = Address.from_bech32(wallet["guardian"])

            totp_code = generate_otp_code(wallet["secret"])
            transaction = create_tx(address, guardian)

            data["code"] = totp_code
            data["transaction"] = transaction

            json_data = json.dumps(data, separators=(',', ':')).encode("utf8")

            self.client.post("/guardian/sign-transaction", headers=headers, data=json_data)

    def sign_tx(self):
        self.sign_tx_from_wallets()


def main():
    path = "../tmp-wallets/"

    wallets = read_wallet_files(path)

    headers = {
        'accept': 'application/json',
        'Content-Type': 'application/json',
    }

    request_url = 'https://testnet-tcs-api.multiversx.com/guardian/sign-transaction'

    for wallet in wallets:
        data = {}

        import pdb; pdb.set_trace()

        address = Address.from_bech32(wallet["address"])
        guardian = Address.from_bech32(wallet["guardian"])

        totp_core = generate_otp_code(wallet["secret"])
        transaction = create_tx(address, guardian)

        data["code"] = totp_core
        data["transaction"] = transaction

        json_data = json.dumps(data, separators=(',', ':')).encode("utf8")

        response = requests.post(request_url, headers=headers, data=json_data)
        print(response)


if __name__ == "__main__":
    main()
