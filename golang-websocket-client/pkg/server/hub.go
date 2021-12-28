package server

import (
	"encoding/json"
	"fmt"
	"log"
)

// Hub maintains the set of active clients and broadcasts messages to the
// clients.
type Hub struct {
	// Registered clients.
	clients    map[string]*Client

	// Inbound messages from the clients.
	broadcast chan Message

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client
	msgIn chan Message//the incoming msg from readpump
}

func NewHub(broadcast chan Message,msgIn chan Message) *Hub {
	return &Hub{
		broadcast:  broadcast,
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[string]*Client),
		msgIn:msgIn,
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			fmt.Println("clients map********h.clients",h.clients)
			fmt.Println("clients map********client",client)
		case client := <-h.unregister:
			if _, ok := h.clients[client.conn.RemoteAddr().String()]; ok {
				delete(h.clients, client.conn.RemoteAddr().String())
				close(client.send)
				h.log("Run", nil, "connection closed")
			}
		case msg:=<-h.broadcast:
			fmt.Println("**************Run()*********in case <-h.broadcast: ",msg)
			msgbytes, err := json.Marshal(msg)
	
	if err != nil {
		log.Println("err in json.Marshal ",err)
	}
			for _,client := range h.clients {
				client.send <-msgbytes
			}
		}
	}
}

// Log print log statement
// In real word I would recommend to use zerolog or any other solution
func (h *Hub) log(f string, err error, msg string) {
	if err != nil {
		fmt.Printf("Error in func: %s, err: %v, msg: %s\n", f, err, msg)
	} else {
		//fmt.Printf("Log in func: %s, %s\n", f, msg)
	}
}

func (h *Hub) CloseConn(){
	for _,client := range h.clients {
		client.conn.Close()
	}
}

func (h *Hub) DeliverToBC(msg Message){
	h.broadcast<-msg
}

func (h *Hub) DeliverToSend(msg Message, cID string){
	msgbytes, err := json.Marshal(msg)
	if err != nil {
		log.Println("err in json.Marshal ",err)
	}
	
	h.clients[cID].send <-msgbytes
	
}