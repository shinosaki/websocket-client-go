package main

import (
	"log"
	"time"

	"github.com/shinosaki/websocket-client-go/websocket"
)

const (
	WEBSOCKET_URL = "wss://echo.websocket.org"
	ATTEMPTS      = 3
	INTERVAL      = 2
)

func main() {
	ws := websocket.NewWebSocketClient(
		// onOpen
		func(ws *websocket.WebSocketClient, isReconnecting bool) {
			if isReconnecting {
				log.Println("Reconnected")
			} else {
				log.Println("Connected")
			}
		},

		// onClose
		func(ws *websocket.WebSocketClient, isReconnecting bool) {
			if isReconnecting {
				log.Println("Reconnecting...")
			} else {
				log.Println("Disconnected")
			}
		},

		// onMessage
		func(ws *websocket.WebSocketClient, payload []byte) {
			log.Println("Received message:", string(payload))
		},
	)

	// Connect to server
	if err := ws.Connect(WEBSOCKET_URL, ATTEMPTS, INTERVAL, false); err != nil {
		log.Println("Failed to connect:", err)
	}

	sendMessage := func() {
		data := map[string]string{"message": "Hello WebSocket"}
		if err := ws.SendJSON(data); err != nil {
			log.Println("Failed to send message:", err)
		}
	}

	// Send a message
	sendMessage()
	time.Sleep(2 * time.Second)

	// Reconnecting
	ws.Reconnect(WEBSOCKET_URL, ATTEMPTS, INTERVAL)
	sendMessage()
	time.Sleep(2 * time.Second)

	ws.Disconnect(false)

	log.Println("Done")
}
