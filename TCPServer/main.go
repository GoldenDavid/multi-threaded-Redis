package main

import (
	"io"
	"log"
	"net"
	"strings"

	"multi-threaded-Redis/Internal/resp"
)

func handleConnection(conn net.Conn) {
	defer conn.Close()
	log.Println("handle conn from =", conn.RemoteAddr())

	parser := resp.NewParser(conn)
	writer := resp.NewWriter(conn)

	for {
		cmd, err := parser.Parse()
		if err != nil {
			if err == io.EOF {
				log.Println("client disconnected: ", conn.RemoteAddr())
			} else {
				log.Println("client read error: ", err)
			}
			return
		}

		log.Printf("command received: %+v\n", cmd)

		// Basic request handling
		if cmd.Type == "array" && len(cmd.Array) > 0 {
			commandName := strings.ToUpper(cmd.Array[0].Bulk)
			
			switch commandName {
			case "PING":
				err = writer.Write(resp.Value{Type: "string", Str: "PONG"})
			default:
				err = writer.Write(resp.Value{Type: "error", Str: "ERR unknown command '" + commandName + "'"})
			}
		} else {
			err = writer.Write(resp.Value{Type: "error", Str: "ERR invalid request"})
		}

		if err != nil {
			log.Println("err write:", err)
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
		conn, err := listener.Accept()
		if err != nil {
			log.Println(err)
			continue
		}

		// create a go routine to handle the connection
		go handleConnection(conn)
	}
}
