# About
Key Server for generating, storing and distributing keys securely.


# Development
### In order to have an optimal development experience you need to have Docker installed.
Create a copy of `.env.example`, rename it to `.env` and update accordingly.

After pulling, go to the root of the project directory, make sure `docker`
is running and then run the commands:
```
docker compose up --build -d
```
This will spin everything up with docker-compose, ie, start the service, database and redis.

For consecutive times, you can run:
`docker compose up -d`


## Health Check

`curl --location 'localhost:9000/key/v1/drm/health'`

### Sample Response
`{
"message": "Service is healthy",
"status": "ok"
}`

## Generate Keys

```
curl --location 'localhost:9000/key/v1/drm/key' \
--header 'Content-Type: application/json' \
--data '{
  "contentId": "12345060889",
  "packageId": "78920123181",
  "quality": ["AUDIO", "HD", "SD", "UHD1", "UHD2"],
  "providerId": "abc1234",
  "drmScheme": ["FP", "PR"]
}'
```

### Expected Response

```
{
    "data": {
        "contentId": "12345060889",
        "packageId": "78920123181",
        "providerId": "abc1234",
        "drmScheme": [
            "FP",
            "PR"
        ],
        "keys": [
            {
                "AUDIO": {
                    "keyId": "1238d45b3a89aeb5f08368fed6ecc2a6",
                    "keyIv": "52301b07ce6e0cddd05810cf0f73468f",
                    "key": "c2859eaf1b1302bdbfadedcc5efbbec7"
                }
            },
            {
                "HD": {
                    "keyId": "99ff21b0837c1c6f658b38afc1304a93",
                    "keyIv": "52301b07ce6e0cddd05810cf0f73468f",
                    "key": "c2859eaf1b1302bdbfadedcc5efbbec7"
                }
            },
            {
                "SD": {
                    "keyId": "f738591ca0f26ef162ecf100d3b6d60d",
                    "keyIv": "52301b07ce6e0cddd05810cf0f73468f",
                    "key": "c2859eaf1b1302bdbfadedcc5efbbec7"
                }
            },
            {
                "UHD1": {
                    "keyId": "e4645dd331e47c6eba0590f3d2761ee6",
                    "keyIv": "52301b07ce6e0cddd05810cf0f73468f",
                    "key": "c2859eaf1b1302bdbfadedcc5efbbec7"
                }
            },
            {
                "UHD2": {
                    "keyId": "1aa542af081cd774721a459e79d10c00",
                    "keyIv": "52301b07ce6e0cddd05810cf0f73468f",
                    "key": "c2859eaf1b1302bdbfadedcc5efbbec7"
                }
            }
        ]
    },
    "message": "Key generated",
    "success": true
}


```


## Stop the Service

`docker compose stop`



