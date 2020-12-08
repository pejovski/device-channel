package main

import (
	"fmt"
	"github.com/labstack/echo"
	"github.com/nats-io/nats.go"
	"log"
	"net/http"
	"os"
	"time"
)

const commandTimeout = time.Second * 3

func main() {
	time.Sleep(time.Second * 2)

	natsURI := os.Getenv("NATS_URI")
	if natsURI == "" {
		natsURI = nats.DefaultURL
	}

	nc, err := nats.Connect(natsURI)
	if err != nil {
		log.Fatalf("Cant connect to %s NATS because %s", natsURI, err)
	}
	defer nc.Close()

	handler := Handler{nats: nc}

	e := echo.New()
	e.GET("/command/:id/:command", handler.command)

	log.Println("Waiting on device commands...")
	e.Logger.Fatal(e.Start(":8005"))
}

type Handler struct {
	nats *nats.Conn
}

func (h *Handler) command(c echo.Context) error {
	id := c.Param("id")
	command := c.Param("command")

	log.Println(fmt.Sprintf("Send command '%s' from device '%s'", command, id))
	_, err := h.nats.Request(id, []byte(command), commandTimeout)
	if err != nil {
		return c.String(http.StatusInternalServerError, fmt.Sprintf("Failed to send command because %s", err))
	}
	return c.String(http.StatusOK, fmt.Sprintf("command '%s' from device '%s' was delivered on %s", command, id, time.Now().Format("2006-01-02 15:04:05")))
}
