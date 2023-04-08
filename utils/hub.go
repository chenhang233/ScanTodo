package utils

import "github.com/gorilla/websocket"

var HubInstance *Hub

type Client struct {
	hub  *Hub
	conn *websocket.Conn
	Send chan []byte
}

type Hub struct {
	Clients       map[*Client]bool
	PrivateClient *Client
	broadcast     chan []byte
	register      chan *Client
	unregister    chan *Client
}

func NewHub() *Hub {
	return &Hub{
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		Clients:    make(map[*Client]bool),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.Clients[client] = true
			h.PrivateClient = client
		case client := <-h.unregister:
			if _, ok := h.Clients[client]; ok {
				delete(h.Clients, client)
				close(client.Send)
			}
		case message := <-h.broadcast:
			for client := range h.Clients {
				select {
				case client.Send <- message:
				default:
					close(client.Send)
					delete(h.Clients, client)
				}
			}
		}
	}
}
