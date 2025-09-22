package main

import (
	"bytes"
	"compress/gzip"
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
	body      []byte
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

		fileName := paths[1]
		directory := os.Args[2]

		if request.startLine.method == "POST" {
			response.parseStartLine("201", "Created")
			fmt.Printf("b %q\n", request.body)
			err := os.WriteFile(directory+"/"+fileName, []byte(request.body), 0644)
			if err != nil {
				conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
				return
			}
			conn.Write([]byte("HTTP/1.1 201 Created\r\n\r\n"))
			return

		}
		f, err := os.ReadFile(directory + "/" + fileName)

		if err != nil {
			conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
			return
		}

		response.parseStartLine("200", "OK")
		response.parseHeaders(*request, "application/octet-stream", strconv.Itoa(len(f)))
		response.parseBody(string(f))
		response := response.Bytes()
		conn.Write([]byte(response))
		return

	}

	if target == "/user-agent" {
		content := request.headers["User-Agent"]
		response.parseStartLine("200", "OK")
		response.parseHeaders(*request, "text/plain", strconv.Itoa(len(content)))
		response.parseBody(content)
		response := response.Bytes()
		conn.Write([]byte(response))
		return

	}

	if paths[0] == "echo" {
		response.parseStartLine("200", "OK")
		response.parseHeaders(*request, "text/plain", strconv.Itoa(len(paths[1])))
		response.parseBody(paths[1])
		response := response.Bytes()
		fmt.Println("res", response)
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
	headersEndIdx := strings.LastIndex(data, "\r\n\r\n")
	headers := strings.Split(data[headersStartIdx:headersEndIdx], "\r\n")
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

	bodyStartIdx := strings.LastIndex(data, "\r\n\r\n") + 4
	bodyEndIdx, err := strconv.Atoi(r.headers["Content-Length"])

	if err != nil {
		return
	}
	r.body = data[bodyStartIdx : bodyStartIdx+bodyEndIdx]
}

func (r *Response) parseStartLine(statusCode string, reason string) {
	r.startLine = ResponseStartLine{
		protocol: "HTTP/1.1",
		status:   statusCode,
		reason:   reason,
	}

}

func (r *Response) parseHeaders(request Request, contentType string, contentLength string) {

	r.headers = map[string]string{
		"Content-Type":   " " + contentType,
		"Content-Length": " " + contentLength,
	}

	if encodings, exists := request.headers["Accept-Encoding"]; exists && strings.Contains(encodings, "gzip") {
		// validEncodings := make([]string, 5)
		// for _, v := range strings.Split(encodings, ",") {
		// 	if strings.TrimSpace(v) != "gzip" {
		// 		continue
		// 	}
		// 	validEncodings = append(validEncodings, "gzip")
		// }
		// fmt.Println("ve", validEncodings)
		r.headers["Content-Encoding"] = " " + "gzip"
	}
}

func (r *Response) parseBody(content string) {

	_, shouldCompress := r.headers["Content-Encoding"]
	if shouldCompress {
		var buffer bytes.Buffer
		compressor := gzip.NewWriter(&buffer)
		_, err := compressor.Write([]byte(content))
		if err != nil {
			panic(err)
		}
		compressor.Close()

		content := buffer.Bytes()

		r.headers["Content-Length"] = " " + strconv.Itoa(len(content))
		r.body = content
		return

	}
	r.body = []byte(content)
}

// func (r Response) String() string {

// 	response := r.startLine.protocol + " " + r.startLine.status + " " + r.startLine.reason + "\r\n"

// 	for k, v := range r.headers {
// 		response += fmt.Sprintf("%s:%s\r\n", k, v)
// 	}

// 	// add the line before the body
// 	response += "\r\n"

// 	response += r.body

// 	return response
// }

func (r Response) Bytes() []byte {
	// build headers as string
	headers := r.startLine.protocol + " " + r.startLine.status + " " + r.startLine.reason + "\r\n"
	for k, v := range r.headers {
		headers += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	headers += "\r\n"

	// combine headers (as bytes) + body (as raw bytes)
	return append([]byte(headers), r.body...)
}
