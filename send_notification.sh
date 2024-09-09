#!/bin/bash
pushkey=$1

# Get the public key from the notification service and save it to pub.pem file
curl http://localhost:8085/api/v1/public -o pub.pem

# Encrypt the app_id and pushkey with the public key
# app_id - name of app from sygnal config
# pushkey - token received by the application from firebase
encrypted_output=$(echo '{
    "app_id":"id.privado.wallet.dev",
    "pushkey":"'"$pushkey"'"
}' | openssl pkeyutl -encrypt -pubin -inkey pub.pem -pkeyopt rsa_padding_mode:oaep -pkeyopt rsa_oaep_md:SHA512 | base64)

# Make http request to the notification service
curl --location --request POST 'http://localhost:8085/api/v1' \
--header 'Content-Type: application/json' \
--data-raw '{
   "message": {"id":"fb112f36-ff8a-414b-bc1d-1fb13b534755"},
   "metadata": {
      "devices": [{
        "ciphertext": "'"$encrypted_output"'",
        "alg": "RSA-OAEP-512"
      }]
   }
}' | jq .