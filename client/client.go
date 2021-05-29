package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

var fileNameToLocation map[string]string
var listFiles []string

const BUFFERSIZE = 1024

var INDEX = "INDEX"
var LIST_ALL = "LIST_ALL"
var REQUEST_SERVER = "REQUEST_SERVER"
var REQUEST_CLIENT = "REQUEST_CLIENT"
var SEND_TO_SERVER = "SEND_TO_SERVER"
var SENT_TO_CLIENT = "SEND_TO_CLIENT"
var ERROR = "ERROR"

func incomingHandler(sock net.Conn) {

	for {
		serverMsg, _ := bufio.NewReader(sock).ReadString('\n')
		splitMessage := strings.Split(serverMsg, "@")
		header := splitMessage[0]

		if header == REQUEST_CLIENT {

			fileName := splitMessage[1]
			receiverSockId := splitMessage[2]
			fmt.Println(serverMsg)
			receiverSockId = strings.Trim(receiverSockId, "\n")
			fileName = strings.Trim(fileName, "\n")

			fmt.Println("Server Requesting " + fileName)

			filePath := fileNameToLocation[fileName]
			bytes, err := ioutil.ReadFile(filePath)

			strBytes := fmt.Sprintf("%v", bytes)
			strBytes = strings.Trim(strBytes, "]")
			strBytes = strings.Trim(strBytes, "[")
			strBytes = strings.Trim(strBytes, " ")

			sendSocketMessage(SEND_TO_SERVER+"@"+receiverSockId+"@"+fileName+"@"+strBytes, sock)

			if err != nil {
				fmt.Println(err)

			}

			fmt.Println("File has been sent")

		} else if header == LIST_ALL {
			files := splitMessage[1]
			files = strings.Trim(files, "\n")
			if files == "" {
				listFiles = []string{}
			} else {
				listFiles = strings.Split(files, "*")
			}

		} else if header == SENT_TO_CLIENT {

			fmt.Println("\nReading")

			FileName := strings.Trim(splitMessage[1], "\n")
			strFile := strings.Trim(splitMessage[2], "\n")

			strBytes := strings.Split(strFile, " ")

			var newBytes []byte

			for b := range strBytes {
				newB, _ := strconv.Atoi(strBytes[b])
				newBytes = append(newBytes, byte(newB))
			}

			ioutil.WriteFile(FileName, newBytes, 0666)

			fmt.Println("\nFile Received.")

		} else if header == ERROR {
			errMessage := splitMessage[1]
			fmt.Println("Error occured. " + errMessage)
		}
	}
}
func outgoingHandler(sock net.Conn) {

	for {
		fmt.Println("\n\n1. Share file\n2. Look for file\n ")
		fmt.Print("Row choice > ")
		choice, _ := bufio.NewReader(os.Stdin).ReadString('\n')

		choice = strings.Trim(choice, "\n")

		if choice == "1" {
			fmt.Print("Enter File path: ")
			filePath, _ := bufio.NewReader(os.Stdin).ReadString('\n')
			filePath = strings.Trim(filePath, "\n")

			if _, err := os.Stat(filePath); err == nil {

				var fileName string

				splitPath := strings.Split(filePath, "/")
				if len(splitPath) == 1 {
					fileName = splitPath[0]
				} else {
					fileName = splitPath[len(splitPath)-1]
				}
				fileNameToLocation[fileName] = filePath
				sendSocketMessage(INDEX+"@"+fileName, sock)
				fmt.Println("File successfuly indexed.")
			} else {
				fmt.Println("File doesn't exist.")
			}

		} else if choice == "2" {

			sendSocketMessage(LIST_ALL+"@", sock)
			fmt.Println("Looking for shared files...")
			time.Sleep(2 * time.Second)

			if len(listFiles) == 0 {
				fmt.Println("\n\nNo files shared so far")
			} else {
				fmt.Println("\n\n\t\t List of files")
				fmt.Println("No.\tFileName")

				for i := 0; i < len(listFiles); i++ {
					fmt.Println(strconv.Itoa(i+1) + "\t" + listFiles[i])
				}

				fmt.Print("Enter file row > ")
				fileIndexStr, _ := bufio.NewReader(os.Stdin).ReadString('\n')
				fileIndex, _ := strconv.Atoi(strings.Trim(fileIndexStr, "\n"))

				if fileIndex >= 1 && fileIndex <= len(listFiles) {
					fileName := strings.Trim(listFiles[fileIndex-1], "\n")

					if fileNameToLocation[fileName] != "" {
						fmt.Println("File is available loacally.")
					} else {

						sendSocketMessage(REQUEST_SERVER+"@"+fileName, sock)
						fmt.Println("Requesting " + fileName)
					}

				}
			}

		}

	}

}

func sendSocketMessage(message string, sock net.Conn) {
	// writer := bufio.NewWriter(sock)
	// writer.WriteString(message + "\n")
	// writer.Flush()

	fmt.Fprintf(sock, message+"\n")

}

func main() {
	fileNameToLocation = make(map[string]string)
	listFiles = []string{}
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
