package main

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
)

type Request struct {
	startLine RequestStartLine
	headers   map[string]string
	body      string
}

type Response struct {
	startLine ResponseStartLine
	headers   map[string]string
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

	for {
		c, err := l.Accept()

		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}

		go handleConn(c)
	}

}

func handleConn(conn net.Conn) {

	buf := make([]byte, 1024)

	_, err := conn.Read(buf)

	if err != nil {
		conn.Write([]byte("HTTP/1.1 404 Not Found"))
		return
	}

	data := string(buf)

	request := &Request{
		headers: make(map[string]string),
	}

	request.build(data)

	response := &Response{
		headers: make(map[string]string),
	}
	target := request.startLine.target
	paths := strings.Split(strings.TrimPrefix(target, "/"), "/")

	if strings.HasPrefix(target, "/files") {
		directory := os.Args[2]

		fileName := paths[1]

		f, err := os.ReadFile(directory + "/" + fileName)

		if err != nil {
			conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
			return
		}

		response.parseHeaders("application/octet-stream", strconv.Itoa(len(f)))
		response.parseBody(string(f))
		response := response.String()
		conn.Write([]byte(response))
		return

	}
	if target == "/user-agent" {
		content := request.headers["User-Agent"]
		response.parseHeaders("text/plain", strconv.Itoa(len(content)))
		response.parseBody(content)
		response := response.String()
		conn.Write([]byte(response))
		return

	}
	if paths[0] == "echo" {
		response.parseHeaders("text/plain", strconv.Itoa(len(paths[1])))
		response.parseBody(paths[1])
		response := response.String()
		conn.Write([]byte(response))
		return
	}

	if target == "/" {

		conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
	} else {
		conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
	}
}

func (r *Request) build(data string) {
	r.parseHeaders(data)
	r.parseStartLine(data)
	r.parseBody(data)
}

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

func (r *Request) parseHeaders(data string) {
	headersStartIdx := strings.Index(data, "\r\n") + 2 // first character of the first header
	fmt.Println("heads", headersStartIdx)
	headersEndIdx := strings.LastIndex(data, "\r\n\r\n")
	fmt.Println("headend", headersEndIdx)
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

func (r *Request) parseBody(data string) {
	r.body = "moda"
}

func (r *Response) parseHeaders(contentType string, contentLength string) {
	r.headers = map[string]string{
		"Content-Type":   " " + contentType,
		"Content-Length": " " + contentLength,
	}
}

func (r *Response) parseBody(content string) {
	r.body = content
}

func (r Response) String() string {
	response := "HTTP/1.1 200 OK\r\n"

	for k, v := range r.headers {
		response += fmt.Sprintf("%s:%s\r\n", k, v)
	}

	// add the line before the body
	response += "\r\n"

	response += r.body

	return response
}
