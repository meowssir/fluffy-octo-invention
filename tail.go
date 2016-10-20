package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

	"golang.org/x/net/websocket"
)

type Server struct {
	clients   map[*Client]string
	broadcast chan []byte
	register  chan *Client
}

func newServer() *Server {
	return &Server{
		broadcast: make(chan []byte),
		register:  make(chan *Client),
		clients:   make(map[*Client]string),
	}
}

type Client struct {
	id     string
	server *Server
	conn   *websocket.Conn
	send   chan []byte
	msg    []byte
}

func randId() string {
	// NOTE: if setting an identifying header upon Dial then this may be redundant.
	rand.Seed(time.Now().UnixNano())
	var runes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	b := make([]rune, 8)
	for i := range b {
		b[i] = runes[rand.Intn(len(runes))]
	}
	return string(b)
}

func (c *Client) readPump() {
	// NOTE: Pump messages from *websocket.Conn to *Server. writePump does the reverse.
	// TODO: defer connection.
	msg := []byte{}
	for {
		err := websocket.Message.Receive(c.conn, &msg)
		fmt.Printf("%v", c.conn.Request().Header)
		if err != nil {
			log.Printf("error: %v", err)
			break
		}
		c.server.broadcast <- msg
	}
}

func (c *Client) writePump(s *Server) {
	// NOTE: Pumps messages from *Server to *websocket.Conn.
	for {
		select {
		case msg := <-c.send:
			// NOTE: added a the key, value pair to the header of the form: 
			// Mongo:[true] to identify who the receiver should be.
			// i.e. the downstream driver client should always send.
			_ = "breakpoint"
			for c := range s.clients {
				// INFO: this operation is not expensive: O(n) where n is c.conn.
				if c.conn.Request().Header.Get("Mongo") != "true" {
					c.conn.Write(msg)
				}
			}
		}
	}
}

func (s *Server) run() {
	for {
		select {
		case client := <-s.register:
			v := randId()
			s.clients[client] = v
			log.Printf("client '%s' registered on Server.", v)
		case message := <-s.broadcast:
			for c := range s.clients {
				c.send <- message
			}
		}
	}
}

var s = newServer()

func webHandler(ws *websocket.Conn) {
	client := &Client{server: s, conn: ws, send: make(chan []byte)}
	go s.run()
	client.server.register <- client
	go client.writePump(s)
	client.readPump()
}

func main() {
	log.Printf("Listening on :12345..")
	http.Handle("/ws", websocket.Handler(webHandler))
	err := http.ListenAndServe(":12345", nil)
	if err != nil {
		panic("ListenAndServe: " + err.Error())
	}
}
