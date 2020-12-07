package main

import (
	"fmt"
	"github.com/labstack/echo"
	"github.com/nats-io/nats.go"
	"log"
	"net/http"
	"time"
)

func main() {
	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		log.Fatal("Cant connect to NATS")
	}
	defer nc.Close()

	handler := Handler{nats: nc}

	e := echo.New()
	e.GET("/command/:id/:command", handler.command)
	e.Logger.Fatal(e.Start(":8005"))
}

type Handler struct {
	nats *nats.Conn
}

func (h *Handler) command(c echo.Context) error {
	id := c.Param("id")
	command := c.Param("command")

	log.Println(fmt.Sprintf("Send command '%s' from device '%s'", command, id))
	_, err := h.nats.Request(id, []byte(command), time.Second*60)
	if err != nil {
		return c.String(http.StatusInternalServerError, fmt.Sprintf("Failed to send command because %s", err))
	}
	return c.String(http.StatusOK, fmt.Sprintf("command '%s' from device '%s' was delivered", command, id))
}
