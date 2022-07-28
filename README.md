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


