package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/sajallimbu/main/pubsub"
	uuid "github.com/satori/go.uuid"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// autoID ... returns a UUID
func autoID() string {
	return uuid.Must(uuid.NewV1(), nil).String()
}

var ps = &pubsub.PubSub{}

func websocketHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	fmt.Println("Client is connected")

	client := pubsub.Client{
		ID:         autoID(),
		Connection: conn,
	}

	ps.AddClient(client)

	for {
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			// if the client is disconnected we need to unsubscribe the user and remove the client from the PubSub repo
			fmt.Printf("Client: %s disconnected\n", client.ID)
			fmt.Println("Removing client and its subscriptions")
			ps.RemoveClient(client)
			fmt.Printf("Total client: %d | Subscriptions: %d\n", len(ps.Clients), len(ps.Subscriptions))
			return
		}
		ps.HandleReceiveMessage(client, messageType, p)

	}
}

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static")
	})

	http.HandleFunc("/ws", websocketHandler)

	fmt.Println("Server is running: http://localhost:3000")
	http.ListenAndServe(":3000", nil)
}
