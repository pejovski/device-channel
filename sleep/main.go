package main

import (
	"bufio"
	"fmt"
	"github.com/nats-io/nats.go"
	"log"
	"net"
	"os"
	"strings"
	"time"
)

func main() {

	natsURI := os.Getenv("NATS_URI")
	if natsURI == "" {
		natsURI = nats.DefaultURL
	}

	log.Println("Start socket server...")
	log.Println("Waiting for a device to request sleeping...")

	ln, _ := net.Listen("tcp", ":8006")

	conn, err := ln.Accept()
	if err != nil {
		log.Println("Failed to accept conn because", err)
	}
	log.Println("Device requested to sleep.")

	idCh := make(chan string, 1)
	done := make(chan bool)

	go func() {
		deviceKnown := false
		for {
			select {
			case <-done:
				return
			default:
				id, err := bufio.NewReader(conn).ReadString('.')
				if err != nil {
					log.Println("Failed to ReadString because", err)
					return
				}
				log.Println("Sleep beacon received:", id)
				if !deviceKnown {
					idCh <- id
				}
				deviceKnown = true
			}
		}
	}()

	id := <-idCh
	close(idCh)

	id = strings.TrimSuffix(id, ".")

	nc, err := nats.Connect(natsURI)
	if err != nil {
		log.Fatal("Cant connect to NATS")
	}
	defer nc.Close()

	log.Printf("Subscribe to device channel %s on behalf of device %s", id, id)
	sub, err := nc.SubscribeSync(id)
	if err != nil {
		log.Fatalf("Failed to SubscribeSync because %s", err)
	}

	msg, err := sub.NextMsg(time.Hour * 10)
	if err != nil {
		log.Fatalf("Failed to NextMsg because %s", err)
	}

	log.Println(fmt.Sprintf("Received a command '%s' from device '%s'. Replay subject %s", string(msg.Data), id, msg.Reply))
	done <- true
	log.Println("Stopped reading from socket.")

	err = sub.Unsubscribe()
	if err != nil {
		log.Fatalf("Failed to unsabscribe because %s", err)
	}
	log.Printf("Unsubscribed from device channel.")

	sub, err = nc.SubscribeSync(fmt.Sprintf("sleep-%s", id))
	if err != nil {
		log.Fatalf("Failed to SubscribeSync because %s", err)
	}
	log.Printf("Subscribed to wakeup sleep-%s", id)

	_, err = fmt.Fprintf(conn, id+".")
	if err != nil {
		log.Fatalln("Failed to send wakeup message to device through socket connection", err)
	}
	log.Println("Send wakeup message to device through socket connection.")

	err = conn.Close()
	if err != nil {
		log.Fatalln("Failed to close socket conn because", err)
	}
	log.Println("Socket connection with the sleeping device closed.")

	log.Println("Waiting on wakeup signal from device.")
	_, err = sub.NextMsg(time.Second * 5)
	if err != nil {
		log.Fatalf("Failed to NextMsg because %s", err)
	}
	log.Println("Wakeup signal from device received.")

	log.Println(fmt.Sprintf("Forward command '%s' from device '%s'", string(msg.Data), id))
	err = nc.PublishMsg(msg)
	if err != nil {
		log.Fatalf("Failed to PublishMsg because %s", err)
	}
}
