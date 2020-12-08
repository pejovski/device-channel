package main

import (
	"bufio"
	"fmt"
	"github.com/nats-io/nats.go"
	"log"
	"net"
	"os"
	"time"
)

func main() {
	time.Sleep(time.Second * 5)

	natsURI := os.Getenv("NATS_URI")
	if natsURI == "" {
		natsURI = nats.DefaultURL
	}

	ssURI := os.Getenv("SLEEP_SERVER_URI")
	if ssURI == "" {
		ssURI = "127.0.0.1:8006"
	}

	id := os.Args[1]

	conn, _ := net.Dial("tcp", ssURI)
	log.Println("Connected to sleep server.")

	ticker := time.NewTicker(time.Second * 5)
	done := make(chan bool)

	go func() {
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				log.Print("Send sleep beacon to sleep server...")
				_, err := fmt.Fprintf(conn, id+".")
				if err != nil {
					log.Println("Error sending beacon because", err)
					break
				}
			}
		}
	}()

	_, err := bufio.NewReader(conn).ReadString('.')
	if err != nil {
		log.Fatalln("Problem reading from server because", err)
	}
	log.Println("Sleep server sent wakeup signal via socket connection.")

	done <- true
	log.Println("Stopped sending beacons.")

	nc, err := nats.Connect(natsURI)
	if err != nil {
		log.Fatalln("Cant connect to NATS")
	}
	defer nc.Close()

	log.Println("Subscribe to device command channel", id)
	_, err = nc.Subscribe(id, func(m *nats.Msg) {
		log.Println(fmt.Sprintf("Received a command '%s' from device '%s'. Replay subject %s", string(m.Data), id, m.Reply))

		err := m.Respond(nil)

		if err != nil {
			log.Println("Failed to respond because", err)
		}
	})
	if err != nil {
		log.Fatal("Cant subscribe to NATS")
	}

	err = nc.Publish(fmt.Sprintf("sleep-%s", id), nil)
	if err != nil {
		log.Fatalf("Failed to PublishMsg because %s", err)
	}
	log.Println("Sent wakeup response to sleep server.")

	select {}
}
