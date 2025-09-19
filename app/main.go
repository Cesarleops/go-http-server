package main

import (
	"fmt"
	"net"
	"os"
	"strings"
)

type Response struct {
	startLine []string
	headers   []string
	body      []string
}

type Request struct {
	startLine []string
	headers   map[string]string
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

	buf := make([]byte, 1024)

	_, err := conn.Read(buf)

	if err != nil {
		conn.Write([]byte("HTTP/1.1 404 Not Found"))
		return
	}

	request := &Request{
		headers: make(map[string]string),
	}

	data := string(buf)

	fmt.Println("Data", data)

	parseRequest(data, request)

	target := request.startLine[1]

	paths := strings.Split(strings.TrimPrefix(target, "/"), "/")

	if paths[0] == "echo" {
		response := fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(paths[1]), paths[1])
		conn.Write([]byte(response))
	} else if target == "/" {
		conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
	} else {
		conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
	}
}

func parseRequest(data string, request *Request) Request {

	request.parseHeaders(data)
	request.parseStartLine(data)

	return *request
}

// func parseResponse(data string, response *Response) {
// 	headers := "Content-Type: text/plain\r\nContent-Length: 3\r\n"
// }

func (r *Request) parseStartLine(data string) {
	startLine := data[:strings.Index(data, "\r\n")]
	startLineSegments := strings.Split(startLine, " ")
	method, target, protocol := startLineSegments[0], startLineSegments[1], startLineSegments[2]
	r.startLine = []string{method, target, protocol}
}

func (r Request) parseHeaders(data string) {

}

func (r Response) parseHeaders(data string) {

}
