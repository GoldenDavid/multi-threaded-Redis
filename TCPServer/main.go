package main

import (
	"log"
	"net"
	"io"
)

func handleConnection(conn net.Conn) {
	log.Println("handle conn from =", conn.RemoteAddr())
	for {
		cmd, err := readCommand(conn)
		log.Println("command:", cmd)
		if err != nil {
			conn.Close()
			log.Println("client disconnected: ", conn.RemoteAddr())
			if err == io.EOF {
				break
			}
		}

		if err = respond(cmd,conn); err != nil {
			log.Println("err write:", err)
		}
	}
}

func readCommand(c net.Conn) (string, error) {
	var buf []byte = make([]byte, 512)
	n, err := c.Read(buf)
	if err != nil {
		return "", err
	}
	return string(buf[:n]), nil
}

func respond(cmd string,c net.Conn) error {
	if _, err := c.Write([]byte(cmd)); err != nil {
		return err
	}
	return nil
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

		// create a go routine to handle the connection
		go handleConnection(conn)
	}

}
