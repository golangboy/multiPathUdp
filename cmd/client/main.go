package main

import (
	"fmt"
	"multiPathUDP"
	"os"
	"strconv"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: go run main.go <port> <addr1> <addr2>")
		return
	}

	port, err := strconv.Atoi(os.Args[1])
	if err != nil {
		fmt.Println("Invalid port number:", os.Args[1])
		return
	}

	addresses := os.Args[2:]

	// 打印参数
	fmt.Printf("Port: %d, Addresses: %v\n", port, addresses)

	c := multiPathUDP.Client{}
	c.ListenRawAnd2Middle(port, addresses)
}
