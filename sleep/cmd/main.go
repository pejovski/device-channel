package main

import (
	"fmt"
	"github.com/nats-io/nats.go"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	id := os.Args[1]

	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		log.Fatal("Cant connect to NATS")
	}
	defer nc.Close()

	log.Println("Subscribe to device", id)
	sub, err := nc.SubscribeSync(id)
	if err != nil {
		log.Fatalf("Failed to SubscribeSync because %s", err)
	}

	msg, err := sub.NextMsg(time.Hour * 10)
	if err != nil {
		log.Fatalf("Failed to NextMsg because %s", err)
	}

	log.Println(fmt.Sprintf("Received a command '%s' from device '%s'. Replay subject %s", string(msg.Data), id, msg.Reply))

	err = sub.Unsubscribe()
	if err != nil {
		log.Fatalf("Failed to unsabscribe because %s", err)
	}

	log.Println("Subscribe to wakeup", id)
	sub, err = nc.SubscribeSync(fmt.Sprintf("sleep-%s", id))
	if err != nil {
		log.Fatalf("Failed to SubscribeSync because %s", err)
	}

	go func() {
		time.Sleep(time.Second * 1)
		log.Println("Wake up device", id)
		_, err = http.Get(fmt.Sprintf("http://localhost:8006/wake-up/%s", id))
		if err != nil {
			log.Fatalf("Failed to wake up device %s because %s", id, err)
		}
	}()

	_, err = sub.NextMsg(time.Second * 2)
	if err != nil {
		log.Fatalf("Failed to NextMsg because %s", err)
	}

	log.Println(fmt.Sprintf("Forward command '%s' from device '%s'", string(msg.Data), id))
	err = nc.PublishMsg(msg)
	if err != nil {
		log.Fatalf("Failed to PublishMsg because %s", err)
	}
}

func command(m *nats.Msg) {
	log.Printf("Received a command: %s\n", string(m.Data))
}
