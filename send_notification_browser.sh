#!/bin/bash

# Get the public key from the notification service and save it to pub.pem file
curl http://localhost:8080/api/v1/public -o pub.pem

# Encrypt with additional unique_id field
# unique_id can be used to get all notifications for a specific user
encrypted_output_with_unique_id=$(echo '{
    "app_id":"billions.web.browser",
    "unique_id":"did:iden3:billions:main:2VmmojYd4TjAdiJqZUGu93iE7B2HZbR94Cv8nTmtP5"
}' | openssl pkeyutl -encrypt -pubin -inkey pub.pem -pkeyopt rsa_padding_mode:oaep -pkeyopt rsa_oaep_md:SHA512 | base64)

# Send encrypted JWE without body
curl --location --request POST 'http://localhost:8080/api/v1' \
--header 'Content-Type: application/json' \
--data-raw '{
   "message": {
            "ciphertext": "fgMb1WEQQTf4-vndsQ5NsV-cl9zqgqZavNV93HWTKl7rRCp_S1bFHZ1EUSq5mOlsT6zB5k8frF4nrBgppiqD2ktrQiymDcIACMEm281HZZgRi0r-qNJpQBzXWfLiWaMEEr8ZDD-mjldeKtXGtazPNOunXFBEqCaEhSjcdr7jqFHgWuEuJDzBOqAnQUhGbTPYbnuFDpMOJXL8znfhfCU3jJz1wvrbdR9mz9sE2YonUXFGag-nekcFdcqPmZyjNJGzvMwYHqYpKZAvC3QQlaS1r98uMg0CMadtLh8PQQzlQaXgq8HW6rdFonNurZELDQ_H_n9t-tBCy_gJwzrnXhq4iNjXDBbPEqRVnuWmDIkThJ8EifwMZsDGXGp-LMWfAn-L5qsVykvXdfK6PU1Q-j0IUpiBdtDo_j6oXERIraT8kU-c21Ww1DmZf9ubz4wJg_Fvy68rSJDuwehQzl9tU-K-V7UbG1ybmA",
            "iv": "_2dk0yGlVcGm-nUi",
            "recipients": [
                {
                    "encrypted_key": "cSznwhjoVAuexmVtnlWIgJJSKviKJld-lfhqLpsEoWYHwszzWu2jr6gb8T5Hs18u8FLmv5_oKRftFCrD0DjtUJVdjMjS7-76Lf2NLyFyR5hfTQKPJ5tifJHRZhDPmD3Rwxqe2AJjuabfL7K7_qPwwLmbClAovOxokSYxpDnFZvJI0W0fbumt2S_7j7nufveWPITx4VoUD__k6T--uljd6uTmbFfi85V5gkWthf03G6MYV-a5UdpeNIcCakQBbx3IPeoRi9gAV48CLEHOpxSG8laAYKENwWZyjMfT7CNhIKVJR5iBcq6LCp-_aDfKZ6sFVsbIuRUh7Qobxrk-lal4vg",
                    "header": {
                        "alg": "RSA-OAEP-256",
                        "kid": "did:iden3:billions:test:2VxnoiNqdMPyHMtUwAEzhnWqXGkEeJpAp4ntTkL8XT#key1"
                    }
                },
                {
                    "encrypted_key": "AyoRNE0e8cbj4vM8L0b6EKZfq-l7lwMwk10q9z9FWcnOjx--rwLAyGSvmS_4_rjjKjtqf1tp7xyH_CqFP2UVx9mnmgyKC6cbCWAuvKtP3eV7GtmKsIBD-f0hMzNAo6h-qhOlyUOIWFUadA8Q3Uf7O6Trfy3yq9V25d13fBTTuN0ta391VPJba9qqw4erfrdWpnlIhoHApCYfW6mejijvQ03lDlOVuY38mTrYgxbA0X_hA_2qPGRzTjchsElWDFeQyzLfAVvCVrrbMDoatM9L6OFP08mK1ltVguMKV0C4IFQGKhXhPAkRz6smBB0OI6CZcyKvQjPekqU6PgRy8mShEA",
                    "header": {
                        "alg": "RSA-OAEP-256",
                        "kid": "did:iden3:polygon:amoy:A6x5sor7zpxUwajVSoHGg8aAhoHNoAW1xFDTPCF49#key1"
                    }
                }
            ],
            "tag": "1jZdSruJ8l_3cC2CoY6vtA",
            "protected": "eyJlbmMiOiJBMjU2R0NNIiwidHlwIjoiYXBwbGljYXRpb24vaWRlbjNjb21tLWVuY3J5cHRlZC1qc29uIn0"
        },
   "metadata": {
      "devices": [{
        "ciphertext": "'"$encrypted_output_with_unique_id"'",
        "alg": "RSA-OAEP-512"
      }]
   }
}' | jq .

# Send iden3comm verification request
curl --location --request POST 'http://localhost:8080/api/v1' \
--header 'Content-Type: application/json' \
--data-raw '{
   "message": {
            "id": "f8aee09d-f592-4fcc-8d2a-8938aa26676c",
            "typ": "application/iden3comm-plain-json",
            "type": "https://iden3-communication.io/authorization/1.0/request",
            "thid": "f8aee09d-f592-4fcc-8d2a-8938aa26676c",
            "from": "did:polygonid:polygon:mumbai:2qFroxB5kwgCxgVrNGUM6EW3khJgCdHHnKTr3VnTcp",
            "body": {
                "callbackUrl": "https://test.com/callback",
                "reason": "age verification",
                "message": "test message",
                "scope": [
                    {
                        "id": 1,
                        "circuitId": "credentialAtomicQueryV3",
                        "params": {
                            "nullifierSessionId": "123443290439234342342423423423423"
                        },
                        "query": {
                            "groupId": 1,
                            "proofType": "BJJSignature",
                            "allowedIssuers": [
                                "*"
                            ],
                            "context": "https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json-ld/kyc-v101.json-ld",
                            "type": "KYCEmployee",
                            "credentialSubject": {
                                "position": {
                                    "$eq": "developer"
                                }
                            }
                        }
                    },
                    {
                        "id": 2,
                        "circuitId": "smallCircuit",
                        "query": {
                            "groupId": 1,
                            "credentialSubject": {
                                "bithdate": {
                                    "$lt": "20010101"
                                }
                            }
                        }
                    },
                    {
                        "id": 3,
                        "circuitId": "credentialAtomicQueryV3",
                        "query": {
                            "allowedIssuers": [
                                "*"
                            ],
                            "context": "https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json-ld/kyc-v101.json-ld",
                            "type": "KYCCountryOfResidenceCredential",
                            "credentialSubject": {
                                "countryCode": {
                                    "$in": [
                                        980,
                                        340
                                    ]
                                }
                            }
                        }
                    }
                ]
            }
        },
   "metadata": {
      "devices": [{
        "ciphertext": "'"$encrypted_output_with_unique_id"'",
        "alg": "RSA-OAEP-512"
      }]
   }
}' | jq .