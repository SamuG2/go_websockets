package main

import (
	"log"
	"math/rand"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
)

var done chan interface{}
var interrupt chan os.Signal

func receiveHandler(connection *websocket.Conn) {
	defer close(done)
	for {
		_, msg, err := connection.ReadMessage()
		if err != nil {
			log.Println("Error in receive:", err)
			return
		}
		log.Printf("Received: %s\n", msg)
	}
}

func main() {
	done = make(chan interface{})    //channel to indicate that the receiverHandler is done
	interrupt = make(chan os.Signal) // channel to listen for interrupt signal to terminate gracefully

	signal.Notify(interrupt, os.Interrupt) //Notify the interrupt channel for SIGINT

	socketUrl := "ws://localhost:9999" + "/socket"
	conn, _, err := websocket.DefaultDialer.Dial(socketUrl, nil)
	if err != nil {
		log.Fatal("Error connecting to Websocket Server:", err)
	}
	defer conn.Close()
	go receiveHandler(conn)
	// Our main loop for the client
	// We send out relevant packets here
	rand.Seed(time.Now().UnixNano())
	for {
		select {
		case <-time.After(time.Duration(1) * time.Millisecond * 10):
			// Send an echo packet every second
			err := conn.WriteMessage(websocket.TextMessage, []byte("Buenos "+strconv.Itoa(rand.Intn(1000))+" días."))
			if err != nil {
				log.Println("Error during writing to websocket:", err)
				return
			}

		case <-interrupt:
			// The  program receives a SIGINT, terminate gracefully
			log.Println("Recieved SIGINT interrupt signal, closing all connections...")

			// Close our websocket connection
			err := conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("Error during closing websocket:", err)
				return
			}
			select {
			case <-done:
				log.Println("Receiver channel closed, exiting...")
			case <-time.After((time.Duration(1) * time.Second)):
				log.Println("Timeout in closing receiving channel. Exiting...")
			}
			return

		}
	}
}
