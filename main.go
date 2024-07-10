package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type Message struct {
	Username string `json:"username"`
	Message  string `json:"message"`
}

var clients = make(map[*websocket.Conn]bool)

var broadcast = make(chan Message)

var wsupgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func main() {
	r := gin.Default()

	r.GET("/", func(ctx *gin.Context) {
		http.ServeFile(ctx.Writer, ctx.Request, "index.html")
	})

	r.GET("/ws", func(ctx *gin.Context) {
		conn, err := wsupgrader.Upgrade(ctx.Writer, ctx.Request, nil)
		if err != nil {
			log.Printf("Error occurred when upgrading")
		}

		clients[conn] = true

		for {
			var message Message
			err := conn.ReadJSON(&message)
			if err != nil {
				log.Printf("Read Error")
				break
			}

			log.Printf("%v", message)

			broadcast <- message
		}
	})

	go handleMessages()

	r.Run(":8001")
}

func handleMessages() {
	for {
		message := <-broadcast
		for client := range clients {
			err := client.WriteJSON(message)
			if err != nil {
				log.Printf("Write Error")
				client.Close()
				delete(clients, client)
			}
		}
	}
}
