package main

import (
	"bufio"
	"fmt"
	"io"
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

func incomingHandler(sock net.Conn) {

	for {
		serverMsg, _ := bufio.NewReader(sock).ReadString('\n')
		splitMessage := strings.Split(serverMsg, "@")
		header := splitMessage[0]

		if header == REQUEST_CLIENT {
			fileName := splitMessage[1]
			receiverSockId := splitMessage[2]
			fmt.Println(serverMsg)
			sendSocketMessage(SEND_TO_SERVER+"@"+receiverSockId, sock)

			filePath := fileNameToLocation[fileName]

			file, err := os.Open(filePath)

			if err != nil {
				fmt.Println(err)

			}

			fileInfo, err := file.Stat()

			if err != nil {
				fmt.Println(err)

			}

			fileSize := int(fileInfo.Size())

			fmt.Println("Sending filename and size")

			sendSocketMessage(strconv.Itoa(fileSize), sock)
			sendSocketMessage(fileName, sock)

			sendBuffer := make([]byte, BUFFERSIZE)
			fmt.Println("Start sending file")

			for {
				_, err = file.Read(sendBuffer)
				if err == io.EOF {
					break
				}
				sock.Write(sendBuffer)
			}
			file.Close()
			fmt.Println("File has been sent")

		} else if header == LIST_ALL {
			files := splitMessage[1]
			listFiles = strings.Split(files, "*")

		} else if header == SENT_TO_CLIENT {
			fmt.Println("Here")

			tempFileSize, _ := bufio.NewReader(sock).ReadString('\n')
			fileSize, _ := strconv.ParseInt(tempFileSize, 10, 64)

			FileName, _ := bufio.NewReader(sock).ReadString('\n')
			FileName = strings.Trim(FileName, "\n")
			newFile, err := os.Create("akayou_" + FileName)

			if err != nil {
				fmt.Println("  Error creating file" + FileName)
			}

			defer newFile.Close()

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
			fmt.Println("\nFile Received.")
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
				fmt.Println("No files shared so far")
			} else {
				fmt.Println("\n\n\t\t List of files")
				fmt.Println("No.\tFileName")

				for i := 0; i < len(listFiles); i++ {
					fmt.Println(strconv.Itoa(i) + "\t" + listFiles[i])
				}

				fmt.Print("Enter file row > ")
				fileIndexStr, _ := bufio.NewReader(os.Stdin).ReadString('\n')
				fileIndex, _ := strconv.Atoi(fileIndexStr)

				if fileIndex >= 0 && fileIndex < len(listFiles) {
					fileName := strings.Trim(listFiles[fileIndex], "\n")

					if fileNameToLocation[fileName] != "" {
						fmt.Println("File is available loacally.")
					} else {
						sendSocketMessage(REQUEST_SERVER+"@"+fileName, sock)

					}

				}
			}

		}

	}

}

func sendSocketMessage(message string, sock net.Conn) {
	writer := bufio.NewWriter(sock)
	writer.WriteString(message + "\n")
	writer.Flush()

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
