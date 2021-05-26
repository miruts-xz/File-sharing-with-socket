package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"sync"
)

var fileNameToLocation map[string]string

const BUFFERSIZE = 1024

var INDEX = "INDEX"
var LIST_ALL = "LIST_ALL"
var REQUEST_SERVER = "REQUEST_SERVER"
var REQUEST_CLIENT = "REQUEST_CLIENT"
var SEND_TO_SERVER = "SEND_TO_SERVER"

func incomingHandler(sock net.Conn) {

	for {
		serverMsg, _ := bufio.NewReader(sock).ReadString('\n')

		fmt.Println(serverMsg)

	}
}
func outgoingHandler(sock net.Conn) {

	for {
		fmt.Println("> ")
		choice, _ := bufio.NewReader(os.Stdin).ReadString('\n')
		sendSocketMessage(choice, sock)

	}

}

func sendSocketMessage(message string, sock net.Conn) {
	writer := bufio.NewWriter(sock)
	writer.WriteString(message)
	writer.Flush()

}

func main() {
	fileNameToLocation = map[string]string{}
	sock, err := net.Dial("tcp", "127.0.0.1:8000")

	if err != nil {
		fmt.Println("An error occured, couldn't connect")
		fmt.Println(err)
		return
	}

	var wg sync.WaitGroup

	wg.Add(2)

	go incomingHandler(sock)
	go outgoingHandler(sock)

	wg.Wait()
}
