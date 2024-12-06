## Fix for the failure after the server is stopped and relaunched (if compiled)

This exporter creates a `netatmo_token.json` file to store the access token, refresh token, and expiration date. If the server stops, it reloads these credentials from the file on relaunch, avoiding the error: `failed to refresh token, status: 400, body: {"error":"invalid_grant"}`

#### Prerequisites

1. Install Go on your machine if not already installed.
2. Go to the [Netatmo Developer Portal](https://dev.netatmo.com/) and:
   - Create a new app.
   - Add token credentials for `read_station` and `read_thermostat`.
   - Copy the generated *Refresh Token*.

#### Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/giuliomagnifico/netatmo_energy_exporter
   
Navigate to the project folder:
    
`cd netatmo_energy_exporter`

Compile the exporter. For example, for ARM (Raspberry Pi):
`GOARCH=arm64 GOOS=linux go build -o netatmo_energy_exporter`   
   
#### Usage

Run the exporter with your credentials. Example:   
```
./netatmo_energy_exporter \
  --client-id 123456789 \
  --client-secret 987654321 \
  --refresh-token "123456789|987654321" \
  --listen 0.0.0.0:2112 \
  --username [your Netatmo username] \
  --password [your Netatmo account password]
```

(`--listen` is optional and can be used to specify a different port)

#### Notes

⚠️ *This fork has been tested with compilation but not with Docker* ⚠️
 
 
-----------


# Netatmo Energy Exporter

This Prometheus exporter works with the netatmo energy API.
It reads the current temperature measurement and set point temperature
and exports it in prometheus readable way alongside with other metrics.
This exporter publishes metrics per room and per modules.

*IMPORTANT*: this exporter works only with netatmo Thermostats and Valves.

## Build Docker Image

The best way to deploy is by creating a docker image by executing:

```shell
docker build -t netatmo_energy_exporter .
```

## Run Docker Container

1. First of all create an App in netatmos developers portal
2. Generate and copy the client id and secret
   * if you're going to use the refresh token, generate one and copy it
3. Run by executing:
    ```shell script
    docker run -d -p 2112:2112 netatmo_energy_exporter \
       --client-id=${CLIENT_ID} --client-secret=${CLIENT_SECRET} \
       --username=${USERNAME} --password=${PASSWORD}
    ```
   or
   ```shell script
   docker run -d -p 2112:2112 netatmo_energy_exporter \
      --client-id=${CLIENT_ID} --client-secret=${CLIENT_SECRET} \
      --refresh-token=${REFRESH_TOKEN}
   ```
   
### Using refresh token

Netatmo has deprecated the ability to use the password credential flow, even though it's still listed.
If you're experiencing any issues while using your username + password combination, try to get the 
refresh token. You should use the following scopes while generating the token:
- read_station
- read_thermostat

### Supported CLI Arguments

--client-id :: netatmo APP client id [*required*]

--client-secret :: netatmo APP client secret [*required*]

--username :: netatmo username [*required*]

--password :: netatmo password [*required*]

--refresh-token :: netatmo refresh token [*required*]

--listen :: address in default go format to listen to (default _0.0.0.0:2112_) [*optional*]
