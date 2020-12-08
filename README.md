# Device Channel POC

#### Create NATS server, run device-channel and sleep-server:


    docker-compose up nats-streaming channel sleep


#### Put device on sleep:


    docker-compose run device ./bin 555


'555' is a deviceID. You can choose other random string.

#### Send a command to device command channel.

Make HTTP GET request form browser:


    http://localhost:8005/command/555/siren


In the example above we demonstrate sending 'SIREN' command to device with ID '555'
