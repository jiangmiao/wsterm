package main

import (
	"log"
	"net/http"

	"github.com/jiangmiao/wsterm"
)

func main() {
	var address = "localhost:9300"
	log.Printf("websocket url ws://%s/ws\n", address)
	http.Handle("/ws", wsterm.Handler)
	http.ListenAndServe(address, nil)
}
