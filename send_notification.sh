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

# Encrypt wit additional unique_id field
# unique_id can be uset to get all notifications for a specific user
encrypted_output_with_unique_id=$(echo '{
    "app_id":"id.privado.wallet.dev",
    "pushkey":"'"$pushkey"'",
    "unique_id":"d2a3c243-8d1d-48f7-8fbe-031f57c1fe4a"
}' | openssl pkeyutl -encrypt -pubin -inkey pub.pem -pkeyopt rsa_padding_mode:oaep -pkeyopt rsa_oaep_md:SHA512 | base64)

curl --location --request POST 'http://localhost:8085/api/v1' \
--header 'Content-Type: application/json' \
--data-raw '{
   "message": {"id":"0130b8b3-26f4-43b4-bc71-768dd60522f0"},
   "metadata": {
      "devices": [{
        "ciphertext": "'"$encrypted_output_with_unique_id"'",
        "alg": "RSA-OAEP-512"
      }]
   }
}' | jq .

curl --location --request POST 'http://localhost:8085/api/v1' \
--header 'Content-Type: application/json' \
--data-raw '{
   "message": {"id":"0130b8b3-26f4-43b4-bc71-768dd60522f0", "body": {"en": "Hello, World!"}},
   "metadata": {
      "devices": [{
        "ciphertext": "'"$encrypted_output_with_unique_id"'",
        "alg": "RSA-OAEP-512"
      }]
   }
}' | jq .