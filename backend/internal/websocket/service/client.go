package service

import (
	"encoding/json"
	"log"
	"time"

	"github.com/Mousa96/chatting-service/internal/websocket/models"
	"github.com/gorilla/websocket"
)

var (
	// This is the duration the server will wait for a Pong message after sending a Ping
	pongWait = 10 * time.Second
	//This is how often the server will send Ping messages to the client
	pingInterval = (pongWait * 9) / 10
	// This is the maximum message size the server will accept
	maxMessageSize = int64(1024 * 1024) // 1MB
)
type EventHandler func(event *models.Event, c *Client) error
type ClientList map[*Client]bool
type Client struct {
	connection *websocket.Conn
	wsService    *WebSocketService
	egress     chan models.Event
	userID     int
	currentConversationWith int
	isActive                bool
}

func NewClient(conn *websocket.Conn, wsService *WebSocketService, userID int) *Client {
	return &Client{
		connection: conn,
		wsService:    wsService,
		egress:     make(chan models.Event, 256),
		userID:     userID,
		currentConversationWith: 0,
		isActive: true,
	}
}

func (c *Client) readMessages() {
	defer func() {
		c.wsService.removeClient(c)
	}()
	if err := c.connection.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
		log.Printf("error setting read deadline: %v", err)
		return
	}

	c.connection.SetReadLimit(maxMessageSize)
	c.connection.SetPongHandler(c.pongHandler)
	for {
		_, payload, err := c.connection.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error reading message: %v", err)
			}
			break
		}
		var request models.Event
		if err := json.Unmarshal(payload, &request); err != nil {
			log.Printf("error unmarshalling message: %v", err)
			continue
		}
		if err := c.wsService.routeEvent(&request, c); err != nil {
			log.Printf("error routing event: %v", err)
			continue
		}
	}
}

func (c *Client) writeMessages() {
	defer func() {
		c.wsService.removeClient(c)
	}()

	ticker := time.NewTicker(pingInterval)

	for {
		select {
		case message, ok := <-c.egress:
			if !ok {
				if err := c.connection.WriteMessage(websocket.CloseMessage, nil); err != nil {
					log.Printf("closing connection :%v", err)
				}
				return
			}


			payload, err := json.Marshal(message)
			if err != nil {
				log.Printf("error marshalling message: %v", err)
			}
			log.Printf("sending message of type: %v and payload: %v", message.Type, string(payload))
			if err := c.connection.WriteMessage(websocket.TextMessage, payload); err != nil {
				log.Printf("error sending message: %v", err)
				return
			}
			log.Println("message sent")

		case <- ticker.C:
			log.Println("sending ping")
			if err := c.connection.WriteMessage(websocket.PingMessage, []byte(``)); err != nil {
				log.Printf("error sending ping: %v", err)
				return
			}
		}
	}
}

func (c *Client) pongHandler(pongMsg string) error {
	log.Println("pong received")
	return c.connection.SetReadDeadline(time.Now().Add(pongWait))
}

