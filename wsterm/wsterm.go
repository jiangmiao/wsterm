package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/jiangmiao/wsterm"
)

func main() {
	port := flag.String("port", "9300", "")
	host := flag.String("host", "localhost", "")
	flag.Parse()
	var address = *host + ":" + *port
	log.Printf("websocket url ws://%s/ws\n", address)
	http.Handle("/ws", wsterm.Handler)
	err := http.ListenAndServe(address, nil)
	if err != nil {
		log.Fatal(err)
	}
}
