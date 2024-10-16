package main

import (
	"fmt"
	"multiPathUDP"
	"os"
	"strconv"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: go run main.go <port> <target>")
		return
	}

	port, err := strconv.Atoi(os.Args[1])
	if err != nil {
		fmt.Println("Invalid port number:", os.Args[1])
		return
	}

	target := os.Args[2]

	// 打印参数
	fmt.Printf("Port: %d, Target: %s\n", port, target)

	s := multiPathUDP.Server{}
	s.ListenMiddleAnd2Target(port, target)
}
