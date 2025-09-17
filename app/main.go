package main

import (
	"fmt"
	"net"
	"os"
)

type Response struct {
	startLine []string
	headers   []string
	body      []string
}

type Request struct {
	startLine []string
	headers   []string
	body      []string
}

func main() {

	l, err := net.Listen("tcp", "0.0.0.0:4221")

	fmt.Println("server", l)
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}

	c, err := l.Accept()
	fmt.Println("c", c)
	if err != nil {
		fmt.Println("Error accepting connection: ", err.Error())
		os.Exit(1)
	}

	handleConn(c)
}

func handleConn(conn net.Conn) {
	fmt.Println("listening")
	conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
}
