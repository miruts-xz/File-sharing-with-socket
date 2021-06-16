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
	"time"
)

var fileNameToLocation map[string]string
var listFiles []string
var loggedIn bool = false

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
const JSON_FILE_PATHS = "filePaths.json"

func incomingHandler(sock net.Conn) {

	for {
		serverMsg, _ := bufio.NewReader(sock).ReadString('\n')
		splitMessage := strings.Split(serverMsg, FLAG_SEPARATOR)
		header := splitMessage[0]

		if header == LOG_IN {

			fmt.Println("\nLog in successful!")
			loggedIn = true

		} else if header == SIGN_UP {
			fmt.Println("\nSign up successful!")
			loggedIn = true
		} else if header == REQUEST_CLIENT {

			fileName := splitMessage[1]
			receiverSockId := splitMessage[2]
			receiverSockId = strings.Trim(receiverSockId, "\n")
			fileName = strings.Trim(fileName, "\n")

			fmt.Println("\nServer Requesting " + fileName)

			filePath := fileNameToLocation[fileName]
			bytes, err := ioutil.ReadFile(filePath)

			if err != nil {
				fmt.Println(err)
				sendSocketMessage(ERROR+FLAG_SEPARATOR+"File can't be found", sock)
				continue
			}

			strBytes := fmt.Sprintf("%v", bytes)
			strBytes = strings.Trim(strBytes, "]")
			strBytes = strings.Trim(strBytes, "[")
			strBytes = strings.Trim(strBytes, " ")

			sendSocketMessage(SEND_TO_SERVER+FLAG_SEPARATOR+receiverSockId+FLAG_SEPARATOR+fileName+FLAG_SEPARATOR+strBytes, sock)

			fmt.Println("\nFile has been sent")

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
			p, _ := os.Getwd()
			fmt.Println(p)
			//fmt.Println(newBytes)
			fmt.Println(p + "\\client\\downloads\\" + FileName)
			ioutil.WriteFile(p+"\\client\\downloads\\"+FileName, newBytes, 0666)

			fmt.Println("\nFile Received.")

		} else if header == ERROR {
			errMessage := splitMessage[1]
			fmt.Println("\nError occured. " + errMessage)
		}
	}
}
func outgoingHandler(sock net.Conn) {

	for {

		if !loggedIn {
			promtBeforeLogin(sock)
		} else {
			promtAfterLogin(sock)
		}

	}

}

func promtBeforeLogin(sock net.Conn) {
	fmt.Println("\n\n1. Log in\n2. Sign up\n ")
	fmt.Print("> ")

	choice, _ := bufio.NewReader(os.Stdin).ReadString('\n')
	choice = strings.Trim(choice, "\n")

	if choice == "1" {

		fmt.Print("username> ")
		username, _ := bufio.NewReader(os.Stdin).ReadString('\n')

		fmt.Print("password> ")
		password, _ := bufio.NewReader(os.Stdin).ReadString('\n')

		username = strings.Trim(username, "\n")
		password = strings.Trim(password, "\n")

		sendSocketMessage(LOG_IN+FLAG_SEPARATOR+username+FLAG_SEPARATOR+password, sock)
		time.Sleep(1 * time.Second)

	} else if choice == "2" {
		fmt.Print("username> ")
		username, _ := bufio.NewReader(os.Stdin).ReadString('\n')

		fmt.Print("password> ")
		password1, _ := bufio.NewReader(os.Stdin).ReadString('\n')

		fmt.Print("Re-enter password> ")
		password2, _ := bufio.NewReader(os.Stdin).ReadString('\n')

		username = strings.Trim(username, "\n")
		password1 = strings.Trim(password1, "\n")
		password2 = strings.Trim(password2, "\n")

		if len(password1) < 4 {
			fmt.Println("Password length should be greater than 3")
			return
		}
		if password1 == password2 {
			sendSocketMessage(SIGN_UP+FLAG_SEPARATOR+username+FLAG_SEPARATOR+password1, sock)
			time.Sleep(1 * time.Second)
		} else {
			fmt.Println("Passwords don't match.")
			return
		}

	}

}

func promtAfterLogin(sock net.Conn) {

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

			splitPath := strings.Split(filePath, "\\")
			if len(splitPath) == 1 {
				fileName = splitPath[0]
			} else {
				fileName = splitPath[len(splitPath)-1]
			}
			fileNameToLocation[fileName] = filePath

			file, _ := json.MarshalIndent(fileNameToLocation, "", " ")
			ioutil.WriteFile(JSON_FILE_PATHS, file, 0644)

			sendSocketMessage(INDEX+FLAG_SEPARATOR+fileName, sock)
			fmt.Println("File successfuly indexed.")
		} else {
			fmt.Println("File doesn't exist.")
		}

	} else if choice == "2" {

		sendSocketMessage(LIST_ALL+FLAG_SEPARATOR, sock)
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
					fmt.Println("File is available locally.")
				} else {

					sendSocketMessage(REQUEST_SERVER+FLAG_SEPARATOR+fileName, sock)
					fmt.Println("Requesting " + fileName)
				}

			}
		}

	}

}

func sendSocketMessage(message string, sock net.Conn) {

	fmt.Fprintf(sock, message+"\n")

}

func main() {
	fileNameToLocation = make(map[string]string)
	listFiles = []string{}

	bytes, error := ioutil.ReadFile(JSON_FILE_PATHS)

	if error != nil {
		fmt.Println("An error occured, couldn't get previously shared files.")
	} else {
		json.Unmarshal(bytes, &fileNameToLocation)
	}

	sock, err := net.Dial("tcp", "192.168.43.15:8000")

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
