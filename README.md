# polygon push gateway
Service to send push notifications through sygnal matrix API for encrypted device info.


# Env variables

### Required:


to start use docker-compose file.

`docker-compose up -d`

**SERVER_HOST** - public URL to polygon push gateway. <br />
**REDIS_URL** - URL to Redis instance. Redis is used for temporary cache of schemas.<br />
**REDIS_PASSWORD** - Redis password.<br />
**GATEWAY_HOST** - URL to sygnal matrix instance <br />
**PRIVATE_KEY** - Encryption key.<br />

### Not required:


**SERVER_PORT** - port to run pgg on. Default: `8085`.<br />
**LOG_LEVEL** - log level. Default `debug`.<br />
**LOG_ENV** - log env. Default `development`.<br />

# Deploy and check
### Deploy
1. Clone this repository.
2. Generate private key for encryption/decryption:
    ```bash
    openssl genrsa -out keypair.pem 4096
    ```
3. Put private key to `.env` file:
    ```bash
    echo "export PRIVATE_KEY='`cat keypair.pem`'" >> .env
    ```
4. Add required env variables (example):
   ```bash
    export GATEWAY_HOST="http://localhost:5000/"
    export SERVER_HOST="http://localhost:8080"
    export SERVER_PORT="8085"
    export REDIS_URL="localhost:6379"
    export REDIS_PASSWORD="secret"
   ```
   GATEWAY_HOST - URL to `sygnal` instance. <br/>
5. Read env variables:
    ```bash
    source .env
    ```
6. Create config for `sygnal` instance by path `.sygnal/sygnal.yaml`:
    ```yaml
   log:
    setup:
     version: 1
     formatters:
      normal:
       format: "%(asctime)s [%(process)d] %(levelname)-5s %(name)s %(message)s"
    access:
     x_forwarded_for: false

   http:
     bind_addresses: ['0.0.0.0']
     port: 5000

   apps:
     polygon.web:
       type: gcm
       api_key: "AAAALiwHn80:...Mh1GcuXWF1dNiTMCcB7ccYR-ocu"
       fcm_options:
         content_available: true
         mutable_content: true
    ```
   **apps** - list of apps. You can add more apps.<br/>
   **apps.polygon.web** - is name of app. You can change it to any name.<br/>
   **apps.app_name.type** - type of app. You can use `gcm` or `fcm` for android devices.<br/>
   **apps.app_name.api_key** - api key for push notifications. You can get it from firebase console.<br/>
   **apps.app_name.fcm_options.content_available** - enable/disable content in a notification message.<br/>
   **apps.app_name.fcm_options.mutable_content** - enables the service extension on the receiving client to handle the image delivered in the payload.<br/>
7. Run docker compose:
    ```bash
    docker-compose up -d
    ```
### Test
1. Run the `send_notification.sh` script and pass the device token as the first parameter:
   ```bash
   ./send_notification.sh ccg01-AeoZLwC7a9HtMOc0...S4_5
   ```
1. Example of success response:
   ```bash
   [
     {
        "device": {
            "ciphertext": "eEvJwHe08uvFqpeU6Ocr2Q5v3+NGjyPCthpIaiJw2/CL7/wAw06yFY0Pn0tLMzVW+ibN/OlH+TzfzEAC8VmzRNWC/98ZYd9t41ihsVwwBD6tYWt/FJE9ZixWhd7TKp7eUC+orTWewbk/JuySMxcOsVtPlKtj+nlqimxBXDc6Vzcgyd35k+EnZ5apQdfwec5cGXCBMV+pRXApACIXlLECl9+dYE7Dv0Zzyas5cC7JUdI9dht13fuElrvoPnacmZtIMefiS4zNxKJI/GvS6tYnoJC76zV3uYex96S5Bdo4ruuQOH7n9SGgqGNtR1H8LpqI0MO02SBfyW5I1CpJOPfeg3HnsZaddOut0A2CmLopUJyJVr9JIFMTNIbD3YoC2VQIbtAKlDcKJLpbqgnz6COBCV7WCtaHUCux7wddA4cvuvdXmUz1dSkBFVJF5ML6iOdC8b50YJpWnEF7h1c1TTJJSfGQge2CrPk5fF14TQQkB+fEjzJBryU9No8quG7FMF1aegeqrScY+C8ELllhubs1lzmJVNzQJnQyIbIB2aPEWa7Uhhdyg1yo/Dfw+Madrkwx9+YYF8LSRrr38Hm6OnwLCPxKlQZ/qDfnJDak7zpfjGAMq9TMkJ3YmIgMO4MljJqskruRFvwWKcRLhOer4NKr3tZv5wxE6KV/U+9SrmHjaR0=",
            "alg": "RSA-OAEP-512"
        },
        "status": "success",
        "reason": ""
     }
   ]
   ```
1. Mobile/web applications should get a notification from the notification service.
