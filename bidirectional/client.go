package main

import (
	"flag"
	"log"
	"net/url"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/websocket"
)

var addr = flag.String("addr", "127.0.0.1:7070", "http service addres")

func main() {
	flag.Parse()
	log.SetFlags(0)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	url := url.URL{
		Scheme: "ws",
		Host:   *addr,
		Path:   "/echo",
	}
	log.Printf("connection to %s", url.String())

	c, _, err := websocket.DefaultDialer.Dial(url.String(), nil)
	if err != nil {
		log.Fatal("dial : ", err)
	}
	defer c.Close()

	done := make(chan struct{})

	c.SetPingHandler(func(data string) error {
		log.Printf("received ping : %s", data)
		if err := c.WriteControl(websocket.PongMessage, []byte{}, time.Time{}); err != nil {
			log.Fatal("send pong : ", err)
		}

		return nil
	})

	go func() {
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read : ", err)
			}
			log.Printf("received message : %s", message)
		}
	}()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return
		// case t := <-ticker.C:
		// 	err := c.WriteMessage(websocket.TextMessage, []byte(t.String()))
		// 	if err != nil {
		// 		log.Println("write : ", err)
		// 		return
		// }
		case <-interrupt:
			log.Println("interrupt")
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("write close : ", err)
				return
			}
		}
		select {
		case <-done:
		case <-time.After(time.Second):
		}
		return
	}
}
