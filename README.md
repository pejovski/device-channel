# Device Channel

Proof of concept for Device Channel & Device Sleep

## Requirements

- docker
- docker-compose

## Install and Run

#### Create NATS server, run device-channel and sleep-server:


    docker-compose up nats channel sleep


#### Put device on sleep:


    docker-compose run device ./bin 555


'555' is a deviceID. You can choose other random string.

Beacons are being sent every 5 seconds. Watch for the logs.

#### Send a command to device command channel.

Make HTTP GET request form browser:


    http://localhost:8005/command/555/siren


In the example above we demonstrate sending 'SIREN' command to device with ID '555'.
