package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"sync"
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
var SENT_TO_CLIENT = "SEND_TO_CLIENT"

func incomingHandler(sock net.Conn) {

	for {

		clientMsg, _ := bufio.NewReader(sock).ReadString('\n')
		splitMessage := strings.Split(clientMsg, "@")
		header := splitMessage[0]

		if header == INDEX { // A client is telling the server the file it wants to share
			fileName := splitMessage[1]
			addFileName(fileName, sock)
			fmt.Println(fileName)
		} else if header == LIST_ALL { // Returns the list of all files available for share

			files := ""
			for key := range fileNames {
				files = key + "*" + files
			}
			files = strings.Trim(files, "*")
			fmt.Println(files)
			sendSocketMessage(LIST_ALL+"@"+files, sock)

		} else if header == REQUEST_SERVER { // A client is requesting server for a file
			fileName := splitMessage[1]

			if _, err := os.Stat("./cache/" + fileName); err == nil { // check if the file is cached on server
				//File is cached on server
				sendFile(fileName, sock)

			} else {
				//File wasn't cached on server so look it up in the FilNames map
				for key := range fileNames {

					fileName = strings.Trim(fileName, "\n")
					if fileName == key {
						sendSocketMessage(REQUEST_CLIENT+"@"+fileName+"@"+socketToId[sock], fileNames[key])
						break
					}
				}

			}
		} else if header == SEND_TO_SERVER {

			fileIsToSocket := sockets[strings.Trim(splitMessage[1], "\n")]

			// Receiving file from provider client

			tempFileSize, _ := bufio.NewReader(sock).ReadString('\n')
			fileSize, _ := strconv.ParseInt(tempFileSize, 10, 64)
			FileName, _ := bufio.NewReader(sock).ReadString('\n')

			readFile(FileName, fileSize, sock)

			// Sending file to requestor client
			sendFile(FileName, fileIsToSocket)

		}
	}

}

func sendFile(fileName string, fileIsToSocket net.Conn) {
	sendSocketMessage(SENT_TO_CLIENT+"@", fileIsToSocket)

	file, err := os.Open("./cache/" + fileName)

	defer file.Close()

	if err != nil {
		fmt.Println(err)

	}

	fileInfo, err := file.Stat()

	if err != nil {
		fmt.Println(err)

	}

	fileSizeToSend := int(fileInfo.Size())

	fmt.Println("Sending filename and size")

	sendSocketMessage(strconv.Itoa(fileSizeToSend), fileIsToSocket)
	sendSocketMessage(fileName, fileIsToSocket)

	sendBuffer := make([]byte, BUFFERSIZE)
	fmt.Println("Start sending file")

	for {
		_, err = file.Read(sendBuffer)
		if err == io.EOF {
			break
		}
		fileIsToSocket.Write(sendBuffer)
	}

	fmt.Println("File has been sent")

}

func readFile(fileName string, fileSize int64, sock net.Conn) {
	fmt.Println("Reading file")
	newFile, err := os.Create("./cache/" + fileName)

	if err != nil {
		fmt.Println("Error creating file" + fileName)
	}

	var receivedBytes int64

	for {
		if (fileSize - receivedBytes) < BUFFERSIZE {
			io.CopyN(newFile, sock, (fileSize - receivedBytes))
			sock.Read(make([]byte, (receivedBytes+BUFFERSIZE)-fileSize))
			break
		}
		io.CopyN(newFile, sock, BUFFERSIZE)
		receivedBytes += BUFFERSIZE
	}
	newFile.Close()

	fmt.Println("File accepted.")
}

func addFileName(fileName string, sock net.Conn) {
	fileNames[strings.Trim(fileName, "\n")] = sock

}

func addCacheFileName(fileName string) {
	cachedFileNames = append(cachedFileNames, fileName)
}

func sendSocketMessage(message string, sock net.Conn) {
	writer := bufio.NewWriter(sock)
	writer.WriteString(message + "\n")
	writer.Flush()

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
