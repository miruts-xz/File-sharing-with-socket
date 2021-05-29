package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
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
var ERROR = "ERROR"

func incomingHandler(sock net.Conn) {

	for {

		clientMsg, _ := bufio.NewReader(sock).ReadString('\n')
		splitMessage := strings.Split(clientMsg, "@")
		header := splitMessage[0]

		if header == INDEX { // A client is telling the server the file it wants to share
			fileName := splitMessage[1]
			fileName = strings.Trim(fileName, "\n")
			addFileName(fileName, sock)
			fmt.Println(fileName + "Added to index")
		} else if header == LIST_ALL { // Returns the list of all files available for share

			files := ""
			for key := range fileNames {
				files = key + "*" + files
			}
			files = strings.Trim(files, "*")
			sendSocketMessage(LIST_ALL+"@"+files, sock)

		} else if header == REQUEST_SERVER { // A client is requesting server for a file
			fileName := splitMessage[1]
			fileName = strings.Trim(fileName, "\n")

			if _, err := os.Stat("./cache/" + fileName); err == nil { // check if the file is cached on server
				//File is cached on server
				bytes, err := ioutil.ReadFile("./cache/" + fileName)

				if err != nil {
					fmt.Println("Error reading file")
				}

				sendFile(fileName, bytes, sock)

			} else {
				//File wasn't cached on server so look it up in the FilNames map
				for key := range fileNames {

					if fileName == key {

						fileOwnerSocket := fileNames[key]

						// one := make([]byte, 1)
						// fileOwnerSocket.SetReadDeadline(time.Now())
						// if _, err := fileOwnerSocket.Read(one); err == io.EOF {

						// 	fileOwnerSocket.Close()
						// 	fmt.Println("Socket is closed")
						// 	//TODO Send an error message to client

						// 	for fn := range fileNames {
						// 		if fileOwnerSocket == fileNames[fn] {
						// 			delete(fileNames, fn)
						// 		}
						// 	}

						// 	fileOwnerSocket.Close()
						// 	delete(sockets, socketToId[fileOwnerSocket])
						// 	delete(socketToId, fileOwnerSocket)

						// 	break

						// }

						// _, err := fileOwnerSocket.Read([]byte{})

						// if err != nil {
						// 	fmt.Println("Error")
						// 	log.Println(err)
						// 	//TODO Send an error message to client

						// 	for fn := range fileNames {
						// 		if fileOwnerSocket == fileNames[fn] {
						// 			delete(fileNames, fn)
						// 		}
						// 	}

						// 	fileOwnerSocket.Close()
						// 	delete(sockets, socketToId[fileOwnerSocket])
						// 	delete(socketToId, fileOwnerSocket)

						// 	break
						// }
						fmt.Println("Sending request to file owner")
						sendSocketMessage(REQUEST_CLIENT+"@"+fileName+"@"+socketToId[sock], fileOwnerSocket)
						break
					}
				}

			}
		} else if header == SEND_TO_SERVER {

			/// ========= Reading file from Sender
			fileIsToSocket := sockets[strings.Trim(splitMessage[1], "\n")]
			FileName := splitMessage[2]
			FileName = strings.Trim(FileName, "\n")

			str := strings.Trim(splitMessage[3], "\n")

			strBytes := strings.Split(str, " ")

			var bytes []byte

			for b := range strBytes {
				newB, _ := strconv.Atoi(strBytes[b])
				bytes = append(bytes, byte(newB))
			}

			//  ======= Caching file

			ioutil.WriteFile("./cache/"+FileName, bytes, 0666)

			//  ======= Sending file
			sendFile(FileName, bytes, fileIsToSocket)

		} else if header == ERROR {
			errMessage := splitMessage[1]
			fmt.Println("Error occured. " + errMessage)
		}
	}

}
func sendFile(fileName string, bytes []byte, fileIsToSocket net.Conn) {

	strBytes := fmt.Sprintf("%v", bytes)
	strBytes = strings.Trim(strBytes, "]")
	strBytes = strings.Trim(strBytes, "[")
	strBytes = strings.Trim(strBytes, " ")
	sendSocketMessage(SENT_TO_CLIENT+"@"+fileName+"@"+strBytes, fileIsToSocket)
}

func addFileName(fileName string, sock net.Conn) {
	fileNames[strings.Trim(fileName, "\n")] = sock

}

func addCacheFileName(fileName string) {
	cachedFileNames = append(cachedFileNames, fileName)
}

func sendSocketMessage(message string, sock net.Conn) {
	// writer := bufio.NewWriter(sock)
	// writer.WriteString(message + "\n")
	// writer.Flush()

	_, err := fmt.Fprintf(sock, message+"\n")

	if err != nil {
		fmt.Println("Socket is not available.")
		for fn := range fileNames {
			if sock == fileNames[fn] {
				delete(fileNames, fn)
			}
		}

		sock.Close()
		delete(sockets, socketToId[sock])
		delete(socketToId, sock)

	}

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
