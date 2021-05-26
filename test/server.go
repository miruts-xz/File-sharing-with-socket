package main

import (
	"bufio"
	"fmt"
	"net"
	"sync"
	"time"
)

const BUFFERSIZE = 1024

var sockets map[string]net.Conn
var socketToId map[net.Conn]string
var fileNames map[string]net.Conn
var cachedFileNames []string

var INDEX = "INDEX"
var LIST_ALL = "LIST_ALL"
var REQUEST_SERVER = "REQUEST_SERVER"
var REQUEST_CLIENT = "REQUEST_CLIENT"
var SEND_TO_SERVER = "SEND_TO_SERVER"

func incomingHandler(sock net.Conn) {

	for {

		clientMsg, _ := bufio.NewReader(sock).ReadString('\n')
		fmt.Println(clientMsg)
		sendSocketMessage("Received message", sock)
	}

}

func addFileName(fileName string, sock net.Conn) {
	fileNames[fileName] = sock
}

func addCacheFileName(fileName string) {
	cachedFileNames = append(cachedFileNames, fileName)
}

func sendSocketMessage(message string, sock net.Conn) {
	fmt.Fprintf(sock, message)
}

func acceptSockets() {
	l, _ := net.Listen("tcp", "127.0.0.1:8000")

	for {
		sock, _ := l.Accept()
		fmt.Println("Accepted")
		time := time.Now()
		sockets[time.String()] = sock
		socketToId[sock] = time.String()

		go incomingHandler(sock)
	}
}

func main() {
	sockets = make(map[string]net.Conn)
	socketToId = make(map[net.Conn]string)
	fileNames = make(map[string]net.Conn)
	cachedFileNames = []string{}

	var wg sync.WaitGroup
	wg.Add(1)

	go acceptSockets()

	wg.Wait()

}
