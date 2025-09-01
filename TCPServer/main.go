package main

import (
	"log"
	"net"
)

func handleConnection(conn net.Conn) {
	defer conn.Close()
	var buf []byte = make([]byte, 1000)
	for {
		_, err := conn.Read(buf)
		if err != nil {
			log.Println("client disconnected: ", err)
			return
		}

		//pretend to process the req
		_, err = conn.Write([]byte("HTTP/1.1 200 OK\r\n" +
			"Content-Type: text/plain\r\n" +
			"Content-Length: 11\r\n" +
			"\r\n" +
			"Hello World"))
		if err != nil {
			log.Println("write error:", err)
			return
		}
	}
}

func main() {
	listener, err := net.Listen("tcp", ":3000")
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Listening at port 3000")

	for {
		//conn == socket == communication channel
		conn, err := listener.Accept()
		if err != nil {
			log.Println(err)
		}

		log.Println("handle conn from =", conn.RemoteAddr())
		// create a go routine to handle the connection
		go handleConnection(conn)
	}

}
