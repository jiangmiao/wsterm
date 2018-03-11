package wsterm

import (
	"log"
	"net/http"
	"sync/atomic"
	"testing"
	"time"

	"golang.org/x/net/websocket"
)

func TestWSTerm(tt *testing.T) {
	go func() {
		http.Handle("/ws", websocket.Handler(WSTerm))
		http.ListenAndServe("127.0.0.1:19300", nil)
	}()
	time.Sleep(200 * time.Millisecond)

	var cmds = []string{"curl", "sudo ping -c 5 127.0.0.1"}
	for _, cmd := range cmds {
		func() {
			var closed int64
			ws, err := websocket.Dial("ws://127.0.0.1:19300/ws", "", "http://localhost")
			if err != nil {
				log.Fatal(err)
			}
			websocket.JSON.Send(ws, Message{
				Type: "exec",
				Data: cmd,
			})
			time.AfterFunc(2*time.Second, func() {
				if atomic.LoadInt64(&closed) == 1 {
					return
				}
				log.Println("kill process", cmd)
				websocket.JSON.Send(ws, Message{
					Type: "stop",
				})
			})
			for {
				var msg Message
				err := websocket.JSON.Receive(ws, &msg)
				if err != nil {
					atomic.StoreInt64(&closed, 1)
					log.Println("error", err)
					break
				}
				log.Println("recv", msg)
			}
		}()
	}

}
