#!/bin/bash

# Get the public key from the notification service and save it to pub.pem file
curl http://localhost:8080/api/v1/public -o pub.pem

# Encrypt with additional unique_id field
# unique_id can be used to get all notifications for a specific user
encrypted_output_with_unique_id=$(echo '{
    "app_id":"billions.web.browser",
    "unique_id":"did:iden3:billions:main:2VmmojYd4TjAdiJqZUGu93iE7B2HZbR94Cv8nTmtP5"
}' | openssl pkeyutl -encrypt -pubin -inkey pub.pem -pkeyopt rsa_padding_mode:oaep -pkeyopt rsa_oaep_md:SHA512 | base64)

curl --location --request POST 'http://localhost:8080/api/v1' \
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

curl --location --request POST 'http://localhost:8080/api/v1' \
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