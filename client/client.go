package client

import (
	"bufio"
	"os"
	"regexp"
	"strings"

	"github.com/brown-csci1380/whatsup/whatsup"

	"fmt"
)

func Start(user string, serverPort string, serverAddr string) {
	// Connect to chat server
	chatConn, err := whatsup.ServerConnect(user, serverAddr, serverPort)
	if err != nil {
		fmt.Printf("unable to connect to server: %s\n", err)
		return
	}

	consoleReader := bufio.NewReader(os.Stdin)
	re := regexp.MustCompile("^@(.*) (.*)")
	fmt.Println("******************************************************************")
	fmt.Println("** To send a message, please type @User followed by the message **")
	fmt.Println("** To find all users, please say list                           **")
	fmt.Println("** To exit, please say bye                                      **")

	for {
		handleReceiveMsg(chatConn)

	Read:
		fmt.Print("> ")
		input, _ := consoleReader.ReadString('\n')

		if strings.HasPrefix(strings.ToLower(input), "bye") {
			fmt.Println("Good bye!")
			whatsup.SendMsg(chatConn, whatsup.WhatsUpMsg{Action: whatsup.DISCONNECT})
			os.Exit(0)
		}

		if strings.HasPrefix(strings.ToLower(input), "list") {
			whatsup.SendMsg(chatConn, whatsup.WhatsUpMsg{Action: whatsup.LIST})
			continue
		}

		match := re.FindStringSubmatch(input)

		if len(match) < 3 {
			fmt.Println("Wrong message format, please type @User followed by the message")
			goto Read
		}

		if match[2] == "" {
			fmt.Println("Message cannot be empty")
			goto Read
		}

		msg := whatsup.WhatsUpMsg{
			Username: match[1],
			Body:     match[2],
			Action:   whatsup.MSG,
		}
		whatsup.SendMsg(chatConn, msg)
	}

	// TODO: Receive input from the user and use the first return value of whatsup.ServerConnect
	// (currently ignored so the stencil will compile) to talk to the server.
}

func handleReceiveMsg(chatConn whatsup.ChatConn) {
	msg, err := whatsup.RecvMsg(chatConn)
	if err != nil {
		fmt.Println(err)
	} else {
		switch msg.Action {
		case whatsup.ERROR, whatsup.LIST:
			fmt.Println(msg.Body)
		case whatsup.MSG:
			fmt.Printf("%s: %s\n", msg.Username, msg.Body)
		}
	}
}
