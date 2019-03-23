package whatsup

import (
	"encoding/gob"
	"fmt"
	"net"
	"sync"
)

type ChatConn struct {
	Enc  *gob.Encoder
	Dec  *gob.Decoder
	Conn net.Conn
}

// ConnectedUser defines the username and chatConn to server of a connected user
type ConnectedUser struct {
	Username string
	ChatConn *ChatConn
}

// ConnectedUsers is a cocurrently safe map of connected users remote IP address to user details
type ConnectedUsers struct {
	users     map[string]ConnectedUser
	usernames map[string]int
	mux       sync.Mutex
}

// ServerConnect Connects a chat client to a chat server
func ServerConnect(username string, serverAddr string, serverPort string) (ChatConn, error) {
	chatConn := ChatConn{}
	fmt.Printf("Connecting to %s:%s\n", serverAddr, serverPort)
	conn, err := net.Dial("tcp", serverAddr+":"+serverPort)
	if err != nil {
		return chatConn, err
	}
	chatConn.Conn = conn
	chatConn.Enc = gob.NewEncoder(conn)
	chatConn.Dec = gob.NewDecoder(conn)

	msg := WhatsUpMsg{Username: username, Action: CONNECT}
	chatConn.Enc.Encode(&msg)

	return chatConn, nil
}

func SendMsg(chatConn ChatConn, msg WhatsUpMsg) {
	chatConn.Enc.Encode(&msg)
}

// RecvMsg Receive next WhatsUpMsg from a ChatConn (blocks)
func RecvMsg(chatConn ChatConn) (WhatsUpMsg, error) {
	var chatMsg WhatsUpMsg
	err := chatConn.Dec.Decode(&chatMsg)
	return chatMsg, err
}

// Add adds a user to the connected user map
func (c *ConnectedUsers) Add(conn ChatConn, username string) {
	c.mux.Lock()
	key := conn.Conn.RemoteAddr().String()
	c.users[key] = ConnectedUser{
		Username: username,
		ChatConn: &conn,
	}
	c.mux.Unlock()
}

// Remove removes a user from the connected user map
func (c *ConnectedUsers) Remove(conn ChatConn) {
	c.mux.Lock()
	// Lock so only one goroutine at a time can access the map c.v.
	defer c.mux.Unlock()
	key := conn.Conn.RemoteAddr().String()
	delete(c.users, key)
}

// Find returns a user from the connected user map for a given chatConn
func (c *ConnectedUsers) Find(conn ChatConn) (ConnectedUser, bool) {
	c.mux.Lock()
	// Lock so only one goroutine at a time can access the map c.v.
	defer c.mux.Unlock()
	key := conn.Conn.RemoteAddr().String()
	res, ok := c.users[key]
	return res, ok
}

// Find returns a user from the connected user map for a given chatConn
func (c *ConnectedUsers) FindByUsername(conn ChatConn) (ConnectedUser, bool) {
	c.mux.Lock()
	// Lock so only one goroutine at a time can access the map c.v.
	defer c.mux.Unlock()
	key := conn.Conn.RemoteAddr().String()
	res, ok := c.users[key]
	return res, ok
}

// List returns a list of connected users
func (c *ConnectedUsers) List() (users []string) {
	c.mux.Lock()
	// Lock so only one goroutine at a time can access the map c.v.
	defer c.mux.Unlock()
	for _, v := range c.users {
		users = append(users, v.Username)
	}
	return users
}

// NewConnectedUsers return a new instance of ConnectedUsers
func NewConnectedUsers() *ConnectedUsers {
	return &ConnectedUsers{
		users: make(map[string]ConnectedUser),
	}
}

func (msg WhatsUpMsg) String() string {
	return fmt.Sprintf("{Username: \"%v\", Body: \"%v\", Action: %v}", msg.Username, msg.Body, msg.Action)
}

func (purpose Purpose) String() string {
	switch purpose {
	case CONNECT:
		return "CONNECT"
	case MSG:
		return "MSG"
	case LIST:
		return "LIST"
	case ERROR:
		return "ERROR"
	case DISCONNECT:
		return "DISCONNECT"
	default:
		return "Unknown Purpose!"
	}
}
