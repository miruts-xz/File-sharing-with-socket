package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
)

var sockets map[string]net.Conn
var socketToUsername map[net.Conn]string
var fileNames map[string]string
var userInfo map[string]string

const LOG_IN = "LOG_IN"
const SIGN_UP = "SIGN_UP"
const INDEX = "INDEX"
const LIST_ALL = "LIST_ALL"
const REQUEST_SERVER = "REQUEST_SERVER"
const REQUEST_CLIENT = "REQUEST_CLIENT"
const SEND_TO_SERVER = "SEND_TO_SERVER"
const SENT_TO_CLIENT = "SEND_TO_CLIENT"
const ERROR = "ERROR"

const FLAG_SEPARATOR = "@"
const JSON_USER_INFO = "userInfo.json"

func incomingHandler(sock net.Conn) {

	for {

		clientMsg, error := bufio.NewReader(sock).ReadString('\n')
		splitMessage := strings.Split(clientMsg, FLAG_SEPARATOR)
		header := splitMessage[0]

		if error != nil {
			cleanUp(sock)
			return
		}

		if header == LOG_IN {
			username := strings.Trim(splitMessage[1], "\n")
			password := strings.Trim(splitMessage[2], "\n")

			realPassword := userInfo[username]

			if realPassword == password {

				loginUser(username, sock)
				sendSocketMessage(LOG_IN+FLAG_SEPARATOR+"Success", sock)
			} else {
				sendSocketMessage(ERROR+FLAG_SEPARATOR+"Incorrect username/password.", sock)
			}

		} else if header == SIGN_UP {
			username := strings.Trim(splitMessage[1], "\n")
			password := strings.Trim(splitMessage[2], "\n")

			if userInfo[username] != "" {
				sendSocketMessage(ERROR+FLAG_SEPARATOR+"Username already taken.", sock)
			} else {
				signupUser(username, password, sock)
				sendSocketMessage(SIGN_UP+FLAG_SEPARATOR+"Success", sock)
			}

		} else if header == INDEX { // A client is telling the server the file it wants to share
			fileName := splitMessage[1]
			fileName = strings.Trim(fileName, "\n")
			addFileName(fileName, sock)
			fmt.Println(fileName + " added to index")
		} else if header == LIST_ALL { // Returns the list of all files available for share

			files := ""
			for key := range fileNames {
				files = key + "*" + files
			}
			files = strings.Trim(files, "*")
			sendSocketMessage(LIST_ALL+FLAG_SEPARATOR+files, sock)

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

				fileOwnerUserName := fileNames[fileName]

				fileOwnerSocket := sockets[fileOwnerUserName]

				if fileOwnerSocket != nil {

					fmt.Println("Sending request to file owner")
					sendSocketMessage(REQUEST_CLIENT+FLAG_SEPARATOR+fileName+FLAG_SEPARATOR+socketToUsername[sock], fileOwnerSocket)

				} else {
					sendSocketMessage(ERROR+FLAG_SEPARATOR+"File owner is not online.", sock)
				}

				// for key := range fileNames {

				// 	if fileName == key {

				// 		fileFound = true
				// 		fileOwnerSocket := fileNames[key]
				// 		fmt.Println("Sending request to file owner")
				// 		sendSocketMessage(REQUEST_CLIENT+FLAG_SEPARATOR+fileName+FLAG_SEPARATOR+socketToUsername[sock], fileOwnerSocket)
				// 		break
				// 	}
				// }

			}
		} else if header == SEND_TO_SERVER {

			/// ========= Reading file from Sender
			fileIsToSocket := sockets[strings.Trim(splitMessage[1], "\n")]

			if fileIsToSocket != nil {

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

			}

		} else if header == ERROR {
			errMessage := splitMessage[1]
			fmt.Println("Error occured. " + errMessage)
		}
	}

}

func loginUser(username string, sock net.Conn) {
	sockets[username] = sock
	socketToUsername[sock] = username
}

func signupUser(username string, password string, sock net.Conn) {
	userInfo[username] = password
	file, _ := json.MarshalIndent(userInfo, "", " ")
	ioutil.WriteFile(JSON_USER_INFO, file, 0644)
	loginUser(username, sock)
}
func sendFile(fileName string, bytes []byte, fileIsToSocket net.Conn) {

	strBytes := fmt.Sprintf("%v", bytes)
	strBytes = strings.Trim(strBytes, "]")
	strBytes = strings.Trim(strBytes, "[")
	strBytes = strings.Trim(strBytes, " ")
	sendSocketMessage(SENT_TO_CLIENT+FLAG_SEPARATOR+fileName+FLAG_SEPARATOR+strBytes, fileIsToSocket)
}

func addFileName(fileName string, sock net.Conn) {

	fileNames[strings.Trim(fileName, "\n")] = socketToUsername[sock]

}

func sendSocketMessage(message string, sock net.Conn) {
	// writer := bufio.NewWriter(sock)
	// writer.WriteString(message + "\n")
	// writer.Flush()

	fmt.Fprintf(sock, message+"\n")

}
func cleanUp(sock net.Conn) {

	sock.Close()
	fmt.Println("Socket is closed")
	//TODO Send an error message to client

	delete(sockets, socketToUsername[sock])
	delete(socketToUsername, sock)

}

func acceptSockets() {
	l, _ := net.Listen("tcp", "127.0.0.1:8000")

	for {
		sock, _ := l.Accept()
		fmt.Println("Accepted")

		go incomingHandler(sock)
	}
}

func main() {
	userInfo = make(map[string]string)
	sockets = make(map[string]net.Conn)
	socketToUsername = make(map[net.Conn]string)
	fileNames = make(map[string]string)

	bytes, error := ioutil.ReadFile(JSON_USER_INFO)

	if error != nil {
		fmt.Println("An error occured, couldn't get previously shared files.")
	} else {
		json.Unmarshal(bytes, &userInfo)
	}

	var wg sync.WaitGroup
	wg.Add(1)

	go acceptSockets()

	wg.Wait()
}
