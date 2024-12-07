# Netatmo Energy Exporter

This Prometheus exporter works with the netatmo energy API.
It reads the current temperature measurement and set point temperature
and exports it in prometheus readable way alongside with other metrics.
This exporter publishes metrics per room and per modules.

*IMPORTANT*: this exporter works only with netatmo Thermostats and Valves.

-----------

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

It will create a new `netatmo_token.json` file containing something like:

  ```json
  {"access_token":"12345|54321","refresh_token":"67890|09876","expiry":"2024-12-07T18:29:27.981698608+01:00"}
```
You can change the year from 2024 to 2025, or to whatever you want, to extend the expiration time.

⚠️ *This fork has been tested with compilation but not with Docker* ⚠️
 
