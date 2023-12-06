package utils

import (
	"ScanTodo/scanLog"
	"fmt"
	"github.com/gorilla/websocket"
	"os"
)

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

func SendToThePrivateClientMsgSuccess(ip string, port uint16, protocol string) string {
	sf := fmt.Sprintf("[成功]:, ip: %s , 端口: %d, 协议: %s, ", ip, port, protocol)
	//if HubInstance.PrivateClient == nil {
	//	fmt.Println("PrivateClient 不存在")
	//	return sf
	//}
	//js, _ := json.Marshal(sf)
	//HubInstance.PrivateClient.Send <- js
	return sf
	//return ""
}

func SendToThePrivateClientMsgError(ip string, port uint16, protocol string, error string) {
	//if HubInstance.PrivateClient == nil {
	//	fmt.Println("PrivateClient 不存在")
	//	return
	//}
	//sf := fmt.Sprintf("[失败]:, ip: %s , 端口: %d, 协议: %s, 失败原因: %s", ip, port, protocol, error)
	//js, _ := json.Marshal(sf)
	//HubInstance.PrivateClient.Send <- js
}

func SendToThePrivateClientCustom(str string) {
	//if HubInstance.PrivateClient == nil {
	//	fmt.Println("PrivateClient 不存在,非WebSocket")
	//	return
	//}
	//sf := fmt.Sprintf("%s", str)
	//js, _ := json.Marshal(sf)
	//HubInstance.PrivateClient.Send <- js
}

func SendSaveIps(l *scanLog.LogConf, str string) {
	logFile, err := os.OpenFile("online_ips.txt", os.O_RDWR|os.O_CREATE|os.O_APPEND, os.ModePerm)
	if err != nil {
		l.Error.Println(str)
		return
	}
	logFile.WriteString(str)
	logFile.WriteString("\r\n")
}
