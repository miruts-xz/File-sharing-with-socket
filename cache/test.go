package main

import (
	"fmt"
	"io/ioutil"
)

func main() {

	bytes, _ := ioutil.ReadFile("./text.pdf")

	strBytes := ""

	// for b := range bytes {
	// 	strBytes += strconv.Itoa(int(bytes[b])) + " "

	// }

	// strBytes = strings.Trim(strBytes, " ")

	fmt.Println(strBytes)
}
