package main

import (
	"io"
	"log"
	"net"

	"multi-threaded-Redis/Internal/database"
	"multi-threaded-Redis/Internal/resp"
)

func handleConnection(conn net.Conn, db *database.Database) {
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

		// Execute command against database
		result := db.Exec(cmd)

		err = writer.Write(result)
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

	// Instantiate the database
	db := database.NewDatabase()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println(err)
			continue
		}

		// create a go routine to handle the connection
		go handleConnection(conn, db)
	}
}
