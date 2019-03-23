package server

import (
	"encoding/gob"

	"github.com/brown-csci1380/whatsup/whatsup"

	"fmt"
	"net"
)

var connectedUsers *whatsup.ConnectedUsers

func handleConnection(conn net.Conn) {
	// TODO: Implement handling messages from a client.
	// You will find whatsup.SendMsg and whatsup.RecvMsg methods useful for
	// serializing and deserializing messages.
	clientChatConn := whatsup.ChatConn{
		Conn: conn,
		Dec:  gob.NewDecoder(conn),
		Enc:  gob.NewEncoder(conn),
	}
	defer whatsup.SendMsg(clientChatConn, whatsup.WhatsUpMsg{Action: whatsup.ERROR, Body: "Server is down"})

	for {
		msg, err := whatsup.RecvMsg(clientChatConn)
		if err != nil {
			fmt.Println(err)
			break
		}
		var reply whatsup.WhatsUpMsg
		switch msg.Action {
		case whatsup.CONNECT:
			if msg.Username == "" {
				reply = whatsup.WhatsUpMsg{Body: "Please provide your name", Action: whatsup.ERROR}
				break
			}
			fmt.Printf("%s connected\n", msg.Username)
			reply = whatsup.WhatsUpMsg{Username: msg.Username, Body: fmt.Sprintf("Hi %s, welcome to WhatsUp!\n", msg.Username), Action: whatsup.MSG}
			connectedUsers.Add(clientChatConn, msg.Username)
			whatsup.SendMsg(clientChatConn, reply)
		case whatsup.MSG:
			// TODO: fix the logic, should be find dest user by user name
			user, ok := connectedUsers.Find(clientChatConn)
			if !ok {
				reply = whatsup.WhatsUpMsg{Body: fmt.Sprintf("%s is unreachable, please try send to another user", msg.Username), Action: whatsup.ERROR}
			}
			reply = msg
			whatsup.SendMsg(*user.ChatConn, reply)
		case whatsup.LIST:
			user, _ := connectedUsers.Find(clientChatConn)
			fmt.Printf("%s requested for connected users\n", user.Username)
			users := connectedUsers.List()
			reply = whatsup.WhatsUpMsg{Body: fmt.Sprintf("Connected Users:\n%v", users), Action: whatsup.LIST}
			whatsup.SendMsg(clientChatConn, reply)
		case whatsup.DISCONNECT:
			fmt.Printf("%s disconnected\n", msg.Username)
			connectedUsers.Remove(clientChatConn)
		default:
			reply = whatsup.WhatsUpMsg{Body: "Sorry, unsupported action", Action: whatsup.ERROR}
			whatsup.SendMsg(clientChatConn, reply)
		}
	}
}

func Start() {
	connectedUsers = whatsup.NewConnectedUsers()
	listen, port, err := whatsup.OpenListener()
	fmt.Printf("Listening on port %d\n", port)

	if err != nil {
		fmt.Println(err)
		return
	}

	for {
		conn, err := listen.Accept() // this blocks until connection or error
		if err != nil {
			fmt.Println(err)
			continue
		}
		go handleConnection(conn)
	}
}
