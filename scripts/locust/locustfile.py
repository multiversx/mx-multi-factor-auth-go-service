import base64

from locust import task, between
from locust.contrib.fasthttp import FastHttpUser
from multiversx_sdk_wallet import Mnemonic
from multiversx_sdk_core import Address


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

    auth_token = "Bearer " + address + ".ZEdWemRHbHVady45MGE2Y2ZiY2IxMTRiN2RkOTU2M2Y5ZTYzYmJmNDYyMzVlNDI3ODUwMjVjNWJhNzJjMzIzZjM2NmJiNzdkMmY2Ljg2NDAwLmUzMA.743dc79fc66112769447a148bd5a403336bf82f7e29d640fbdd8cf3d5e12637256992155d7138c3e04f3ff17924042a5496e6497ca51072eca204ceaff469b06"

    header = {'Authorization': auth_token, 'Content-Type': 'application/json'}
    return header


class WebsiteUser(FastHttpUser):
    wait_time = between(1, 2)

    @task(1)
    def get_index(self):
        header = generateMockAuthorizationToken()
        
        self.client.post("/guardian/register", headers=header, json={'tag':'darius'})
