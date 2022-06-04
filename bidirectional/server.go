package main

import (
	"flag"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{} // use default options

var addr = flag.String("addr", "127.0.0.1:7070", "http service addres")

func main() {
	flag.Parse()
	log.SetFlags(0)
	http.HandleFunc("/echo", echo)
	err := http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Fatal("ListenAndServer : ", err)
	}
}

func echo(w http.ResponseWriter, r *http.Request) {
	e, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("upgrade : ", err)
		return
	}
	defer e.Close()

	done := make(chan struct{})

	// websocket only are able use no more than one goroutine to calls
	// write methods and read methods

	go receiver(e, done) //goroutine for receive message -> read methods
	go sender(e, done)   // goroutine for send message -> write methods

	<-done
	log.Println("websocket handler is done!!!")
}

func receiver(ws *websocket.Conn, done chan struct{}) {
	defer func() {
		ws.Close()
	}() // close connection websocket

	// setting handler
	ws.SetPongHandler(func(string) error {
		log.Println("received pong")
		return nil
	})

	// looping for
	for {
		_, message, err := ws.ReadMessage()
		if err != nil {
			log.Println("read : ", err)
			break
		}
		log.Println("received message : ", message)
	}

	close(done) // close goroutine channel
}

func sender(ws *websocket.Conn, done chan struct{}) {
	defer func() {
		ws.Close()
	}() // close connection websocket

	// setting message ticker
	messageTicker := time.NewTicker(2 * time.Second)
	defer messageTicker.Stop() // close goroutine channel

	// setting ping ticker. Ping ticker should longer than message ticker
	pingTicker := time.NewTicker(6 * time.Second)
	defer pingTicker.Stop() // close goroutine channel

	counter := 0

breakLoop:
	for {
		select {
		// sending goroutine channel to messageTicker
		case <-messageTicker.C:
			data := "hello world !!!"
			if err := ws.WriteMessage(websocket.TextMessage, []byte(data)); err != nil {
				log.Println("write error : ", err)
				return
			}
			// stopping breakLoop
			if counter > 20 {
				break breakLoop
			}
			counter++

			// sending goroutine channel to pingTicker
		case <-pingTicker.C:
			if err := ws.WriteControl(websocket.PingMessage, []byte("ping message"), time.Time{}); err != nil {
				log.Println("ping error: ", err)
				return
			}
			// close channel
		case <-done:
			return
		}
	}

	// ensuring close the websocket if there are no more connections or no more message
	ws.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	close(done) // close goroutine channel
}
