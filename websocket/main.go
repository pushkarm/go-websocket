package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"strconv"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
)

// ClientList is a map used to help manage a map of clients
type ClientList map[*Client]bool

// Client is a websocket client, basically a frontend visitor
type Client struct {
	// the websocket connection
	connection net.Conn
	name       string
}

type UserMessage struct {
	EventType string `json:"type"`
	Username  string `json:"username"`
	Content   string `json:"content"`
}

type ContentChangeMessage struct {
	EventType string            `json:"type"`
	Data      map[string]string `json:"data"`
}

func NewClient(conn net.Conn, randomNumber string) *Client {
	return &Client{
		connection: conn,
		name:       randomNumber,
	}
}

func addClient(clients ClientList, conn net.Conn, randomNumber string) {
	fmt.Println("New client:" + randomNumber)
	client := NewClient(conn, randomNumber)
	clients[client] = true
}

func sendMessage(clients ClientList, msg []byte, op ws.OpCode) {

	fmt.Println("Received data:  " + string(msg))
	var userMessage UserMessage
	err := json.Unmarshal(msg, &userMessage)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println("event type: "+userMessage.EventType+userMessage.Username, " ", userMessage.Content)
	for client := range clients {
		//fmt.Println(client.name)
		if userMessage.EventType == "contentchange" {
			data := map[string]string{"editorContent": userMessage.Content}
			c := ContentChangeMessage{EventType: "contentchange", Data: data}

			dataByte, _ := json.Marshal(c)
			err := wsutil.WriteServerMessage(client.connection, ws.OpText, dataByte)
			if err != nil {
				fmt.Println("Error sending data: " + err.Error())
				//fmt.Println("Client disconnected")
				delete(clients, client)
				return
			}
		}
	}
}

func main() {

	clients := make(ClientList)

	fmt.Println("Server started, waiting for connection from client")
	http.ListenAndServe(":8000", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Client connected")
		conn, _, _, err := ws.UpgradeHTTP(r, w)
		if err != nil {
			fmt.Println("Error starting socket server: " + err.Error())
		}
		defer conn.Close()

		randomNumber := strconv.Itoa(rand.Intn(100))
		addClient(clients, conn, randomNumber)
		//go func() {
		//defer conn.Close()
		for {
			msg, op, err := wsutil.ReadClientData(conn)
			if err != nil {
				fmt.Println("Error receiving data: " + err.Error())
				fmt.Println("Client disconnected")
				return
			}
			sendMessage(clients, msg, op)
		}
		//}()
	}))
}
