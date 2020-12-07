package main

import (
	"fmt"
	"github.com/labstack/echo"
	"github.com/nats-io/nats.go"
	"log"
	"net/http"
	"sync"
	"time"
)

var doOnce sync.Once

func main() {
	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		log.Fatal("Cant connect to NATS")
	}
	defer nc.Close()

	handler := Handler{nats: nc}

	e := echo.New()
	e.GET("/wake-up/:id", handler.wakeUp)
	e.Logger.Fatal(e.Start(":8006"))
}

type Handler struct {
	nats *nats.Conn
}

func (h *Handler) wakeUp(c echo.Context) error {
	id := c.Param("id")

	doOnce.Do(func() {
		go func() {
			log.Println("Subscribe to device", id)
			_, err := h.nats.Subscribe(id, func(m *nats.Msg) {
				log.Println(fmt.Sprintf("Received a command '%s' from device '%s'. Replay subject %s", string(m.Data), id, m.Reply))

				time.Sleep(time.Second * 2)
				err := m.Respond(nil)

				if err != nil {
					log.Println("Failed to respond because", err)
				}
			})
			if err != nil {
				log.Fatal("Cant subscribe to NATS")
			}
		}()
	})

	err := h.nats.Publish(fmt.Sprintf("sleep-%s", id), []byte(""))
	if err != nil {
		return c.String(http.StatusInternalServerError, fmt.Sprintf("Failed to send command because %s", err))
	}

	log.Printf("Device %s is woken up", id)
	return c.String(http.StatusOK, fmt.Sprintf("Device %s is woken up", id))
}
