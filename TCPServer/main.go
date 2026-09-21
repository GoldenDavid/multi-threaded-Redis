package main

import (
	"io"
	"log"
	"net"

	"multi-threaded-Redis/Internal/database"
	"multi-threaded-Redis/Internal/resp"
)

type ConnectionJob struct {
	client net.Conn
}

type CommandJob struct {
	client net.Conn
	cmd    resp.Value
}

type ResponseJob struct {
	client net.Conn
	result resp.Value
}

func parserWorker(connQueue <-chan ConnectionJob, cmdQueue chan<- CommandJob) {
	for job := range connQueue {
		handleConnection(job.client, cmdQueue)
	}
}

func handleConnection(conn net.Conn, cmdQueue chan<- CommandJob) {
	defer conn.Close()
	log.Println("handle conn from =", conn.RemoteAddr())

	parser := resp.NewParser(conn)

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
		
		// Push parsed command to the Command Queue
		cmdQueue <- CommandJob{client: conn, cmd: cmd}
	}
}

func responseWorker(resQueue <-chan ResponseJob) {
	for job := range resQueue {
		writer := resp.NewWriter(job.client)
		err := writer.Write(job.result)
		if err != nil {
			log.Println("err write to client", job.client.RemoteAddr(), ":", err)
		}
	}
}

func dbExecutor(db *database.Database, cmdQueue <-chan CommandJob, resQueue chan<- ResponseJob) {
	log.Println("DB Executor started")
	for job := range cmdQueue {
		// Execute command sequentially in a single goroutine
		result := db.Exec(job.cmd)
		
		// Push the result to the Response Queue
		resQueue <- ResponseJob{client: job.client, result: result}
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

	// Create queues for inter-thread communication
	connQueue := make(chan ConnectionJob, 1000)
	cmdQueue := make(chan CommandJob, 1000)
	resQueue := make(chan ResponseJob, 1000)

	// 1. Start the single-threaded Database Executor
	go dbExecutor(db, cmdQueue, resQueue)

	// 2. Start a pool of Response Workers
	numResponseWorkers := 4
	for i := 0; i < numResponseWorkers; i++ {
		go responseWorker(resQueue)
	}

	// 3. Start a pool of Parser Workers (Connection Pool)
	numParserWorkers := 4
	for i := 0; i < numParserWorkers; i++ {
		go parserWorker(connQueue, cmdQueue)
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Accept error:", err)
			continue
		}

		// 4. Dispatch the connection to the parser worker pool
		connQueue <- ConnectionJob{client: conn}
	}
}
