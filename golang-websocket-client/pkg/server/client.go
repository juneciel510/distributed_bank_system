// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package server

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// Client is a middleman between the websocket connection and the hub.
type Client struct {
	hub *Hub

	// The websocket connection.
	conn *websocket.Conn

	// Buffered channel of outbound messages.
	send chan []byte
}

type Message struct {
	Command   string
	Parameter interface{}
}

// readPump pumps messages from the websocket connection to the hub.
//
// The application runs readPump in a per-connection goroutine. The application
// ensures that there is at most one reader on a connection by executing all
// reads from this goroutine.
func (c *Client) readPump() {
	
	defer func() {
		c.conn.Close()
	}()
	
	for {
		var message Message
		err := c.conn.ReadJSON(&message)
		if err != nil {
			 if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
			 	c.hub.log("readPump", err, "reading error")
			 }
			log.Println("error in readPump",err)
			// panic(err)
		
			break
		}

		if message.Command!="HEARTBEAT"{
			log.Println("************************in readPump(),message: ",message)
		}
		if message.Command=="CLIENT"{
			params := message.Parameter.(map[string]interface{})
			cID := params["ClientID"].(string)
			c.hub.clients[cID]=c
			fmt.Println("***the client has been registed,c.hub.clients:",c.hub.clients)
		
		}
		c.hub.msgIn <- message
	}
}

// writePump pumps messages from the hub to the websocket connection.
//
// A goroutine running writePump is started for each connection. The
// application ensures that there is at most one writer to a connection by
// executing all writes from this goroutine.
func (c *Client) writePump() {
	
	defer func() {

	}()
	for {
		select {
		case message, ok := <-c.send:
		
			if !ok {
				// The hub closed the channel.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			log.Println("write message: ",string(message))
			w.Write(message)

			if err := w.Close(); err != nil {
				log.Println("writePump err",err)
				return
			}
		}
	}
}

// serveWs handles websocket requests from the all clients and other servers.
func ServeWs(hub *Hub, w http.ResponseWriter, r *http.Request) {
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		hub.log("ServeWs", err, "cannot handle websocket request")
		return
	}
	client := &Client{hub: hub, conn: conn, send: make(chan []byte, 256)}

	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go client.writePump()
	go client.readPump()
}

