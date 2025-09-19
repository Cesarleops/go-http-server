package main

import (
	"fmt"
	"net"
	"os"
	"strings"
)

type Request struct {
	startLine RequestStartLine
	headers   map[string]string
	body      string
}

type Response struct {
	startLine ResponseStartLine
	headers   []string
	body      string
}

type RequestStartLine struct {
	method   string
	target   string
	protocol string
}

type ResponseStartLine struct {
	protocol string
	status   string
	reason   string
}

func main() {

	l, err := net.Listen("tcp", "0.0.0.0:4221")

	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}

	c, err := l.Accept()

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

	target := request.startLine.target
	paths := strings.Split(strings.TrimPrefix(target, "/"), "/")

	if target == "/user-agent" {
		agent := request.headers["User-Agent"]
		fmt.Printf("%q\n", agent)
		response := fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(agent), request.headers["User-Agent"])
		conn.Write([]byte(response))
	}
	if paths[0] == "echo" {
		response := fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(paths[1]), paths[1])
		conn.Write([]byte(response))
	}
	if target == "/" {
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

// func parseResponse(body string, response *Response) {

// 	startLine := "HTTP/1.1 200 OK\r\n"
// 	headers := map[string]string{
// 		"Content-Type":   "Application/json",
// 		"Content-Length": string(len(body)),
// 	}
// }

func (r *Request) parseStartLine(data string) {
	startLine := data[:strings.Index(data, "\r\n")]
	startLineSegments := strings.Split(startLine, " ")
	method, target, protocol := startLineSegments[0], startLineSegments[1], startLineSegments[2]
	r.startLine = RequestStartLine{
		target:   target,
		method:   method,
		protocol: protocol,
	}
}

func (r Request) parseHeaders(data string) {
	headersStartIdx := strings.Index(data, "\r\n") + 2 // first character of the first header
	headersEndIdx := strings.LastIndex(data, "\r\n\r\n")
	headers := strings.Split(data[headersStartIdx:headersEndIdx], "\r\n")
	fmt.Printf("%q\n", headers)
	for _, header := range headers {
		kv := strings.SplitN(string(header), ":", 2)
		if len(kv) < 2 {
			continue
		}
		k := kv[0]
		v := strings.TrimSpace(kv[1])
		r.headers[k] = v
	}

}

func (r Request) parseBody(data string) {
	headersEndIdx := strings.LastIndex(data, "\r\n\r\n") + 4
	body := data[headersEndIdx:]
	r.body = body
}
