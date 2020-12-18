package pubsub

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/gorilla/websocket"
)

// PubSub ... holds the array of the client and subscription
type PubSub struct {
	Clients       []Client
	Subscriptions []Subscriptions
}

// Client ... a client struct
type Client struct {
	ID         string
	Connection *websocket.Conn
}

// Message ... our message structure
type Message struct {
	Action  string          `json:"action"`
	Topic   string          `json:"topic"`
	Message json.RawMessage `json:"message"`
}

// Subscriptions ... our subscription model
type Subscriptions struct {
	Topic  string
	Client *Client
}

// type of action by the user
var (
	PUBLISH     = "publish"
	SUBSCRIBE   = "subscribe"
	UNSUBSCRIBE = "unsubscribe"
)

// AddClient ... imported functions should always start with UpperCase letters
func (ps *PubSub) AddClient(client Client) *PubSub {

	ps.Clients = append(ps.Clients, client)
	// fmt.Println("Adding new client to the list: ", client.ID)

	payload := []byte("Hello client ID: " + client.ID)
	client.Connection.WriteMessage(1, payload)
	return ps
}

// RemoveClient ... remove the client subscriptions and the client itself from the client list
func (ps *PubSub) RemoveClient(client Client) *PubSub {

	// remove all subscriptions by this client
	for index, sub := range ps.Subscriptions {
		if client.ID == sub.Client.ID {
			ps.Subscriptions = append(ps.Subscriptions[:index], ps.Subscriptions[index+1:]...)
		}
	}

	// remove client from the client list
	for index, cl := range ps.Clients {
		if cl.ID == client.ID {
			ps.Clients = append(ps.Clients[:index], ps.Clients[index+1:]...)
		}
	}
	return ps
}

// HandleReceiveMessage ... handles the received message
func (ps *PubSub) HandleReceiveMessage(client Client, messageType int, payload []byte) *PubSub {

	m := Message{}

	err := json.Unmarshal(payload, &m)
	if err != nil {
		log.Println("JSON marshal error")
		return ps
	}

	switch m.Action {
	case PUBLISH:
		ps.Publish(m.Topic, m.Message, &client)
		break

	case SUBSCRIBE:
		ps.Subscribe(&client, m.Topic)
		fmt.Printf("New subscriber added to topic: %s | Total subscribers: %d\n", m.Topic, len(ps.Subscriptions))
		break

	case UNSUBSCRIBE:
		ps.Unsubscribe(&client, m.Topic)
		break

	default:
		break
	}

	return ps
}

// Subscribe ... takes a client and a topic, add the new client to the subscription and return the pointer to the struct
func (ps *PubSub) Subscribe(client *Client, topic string) *PubSub {

	clientSubs := ps.GetSubscriptions(topic, client)

	if len(clientSubs) > 0 {
		// client is already subscribed to this topic so dont do anything
		return ps
	}

	newSubscription := Subscriptions{
		Topic:  topic,
		Client: client,
	}

	ps.Subscriptions = append(ps.Subscriptions, newSubscription)

	return ps
}

// Unsubscribe ... unsubscribes user from the subscription
func (ps *PubSub) Unsubscribe(client *Client, topic string) *PubSub {

	for index, sub := range ps.Subscriptions {
		if sub.Client.ID == client.ID && sub.Topic == topic {
			// Found the subscription, remove it
			ps.Subscriptions = append(ps.Subscriptions[:index], ps.Subscriptions[index+1:]...)
			fmt.Printf("Unsubscribed successfully. Client ID: %s, Topic: %s\n", client.ID, topic)
		}
	}
	return ps
}

// Send ... a function to send messages
func (client *Client) Send(message []byte) error {
	return client.Connection.WriteMessage(1, message)
}

// GetSubscriptions ... returns an array of subscriptions
func (ps *PubSub) GetSubscriptions(topic string, client *Client) []Subscriptions {

	var subscriptionList []Subscriptions

	// check if the client has any matching subscriptions
	for _, subscription := range ps.Subscriptions {
		// check if the client is nil
		if client != nil {
			if subscription.Client.ID == client.ID && subscription.Topic == topic {
				// add the matched subscription to the list for future use
				subscriptionList = append(subscriptionList, subscription)
			}
		} else {
			if subscription.Topic == topic {
				// add the subscription even if only the topic matches
				subscriptionList = append(subscriptionList, subscription)
			}
		}
	}
	return subscriptionList
}

// Publish ... publishes the message to clients
func (ps *PubSub) Publish(topic string, message []byte, excludeClient *Client) {

	subscriptions := ps.GetSubscriptions(topic, nil)
	for _, sub := range subscriptions {
		if excludeClient.ID != sub.Client.ID {
			fmt.Printf("\nSending to client id %s.Message: %s\n", sub.Client.ID, message)
			sub.Client.Send(message)
		}
	}
}
