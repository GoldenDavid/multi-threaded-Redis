package main

import (
	"io"
	"log"
	"net"

	"multi-threaded-Redis/Internal/aof"
	"multi-threaded-Redis/Internal/core/io_multiplexing"
	"multi-threaded-Redis/Internal/database"
	"multi-threaded-Redis/Internal/resp"
)

type ClientContext struct {
	fd     int
	conn   net.Conn
	parser *resp.Parser
}

type ReadEventJob struct {
	ctx      *ClientContext
	fallback bool
}

type CommandJob struct {
	client net.Conn
	cmd    resp.Value
}

type ResponseJob struct {
	client net.Conn
	result resp.Value
}

func parserWorker(connQueue <-chan ReadEventJob, cmdQueue chan<- CommandJob) {
	for job := range connQueue {
		if job.fallback {
			// Fallback: block and read continuously
			for {
				cmd, err := job.ctx.parser.Parse()
				if err != nil {
					if err == io.EOF {
						log.Println("client disconnected:", job.ctx.conn.RemoteAddr())
					} else {
						log.Println("client read error:", err)
					}
					job.ctx.conn.Close()
					break
				}
				log.Printf("command received: %+v\n", cmd)
				cmdQueue <- CommandJob{client: job.ctx.conn, cmd: cmd}
			}
		} else {
			// Event-driven: read available data then return to pool
			for {
				cmd, err := job.ctx.parser.Parse()
				if err != nil {
					if err == io.EOF {
						log.Println("client disconnected:", job.ctx.conn.RemoteAddr())
					} else {
						log.Println("client read error:", err)
					}
					job.ctx.conn.Close()
					break
				}
				log.Printf("command received: %+v\n", cmd)
				cmdQueue <- CommandJob{client: job.ctx.conn, cmd: cmd}
				
				if !job.ctx.parser.HasMoreData() {
					break
				}
			}
		}
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

	// Initialize AOF
	aofPersister, err := aof.NewAof("appendonly.aof")
	if err != nil {
		log.Println("AOF init error:", err)
	} else {
		defer aofPersister.Close()

		log.Println("Replaying AOF...")
		aofPersister.Read(func(value resp.Value) {
			db.Exec(value)
		})
		log.Println("AOF replay complete.")

		// Bind AOF to Database to log future writes
		db.SetAof(aofPersister)
	}

	// Create queues for inter-thread communication
	connQueue := make(chan ReadEventJob, 1000)
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

	multiplexer, err := io_multiplexing.CreateIOMultiplexer()
	if err != nil {
		log.Println("Multiplexer not supported, using blocking accept:", err)
		for {
			conn, err := listener.Accept()
			if err != nil {
				log.Println("Accept error:", err)
				continue
			}

			// 4. Dispatch the connection to the parser worker pool
			clientCtx := &ClientContext{
				conn:   conn,
				parser: resp.NewParser(conn),
			}
			connQueue <- ReadEventJob{ctx: clientCtx, fallback: true}
		}
	} else {
		defer multiplexer.Close()
		log.Println("Using IO Multiplexing for connections")

		tcpListener := listener.(*net.TCPListener)
		file, err := tcpListener.File()
		if err != nil {
			log.Fatal(err)
		}
		serverFd := int(file.Fd())

		err = multiplexer.Monitor(io_multiplexing.Event{Fd: serverFd, Op: io_multiplexing.OpRead})
		if err != nil {
			log.Fatal(err)
		}

		clients := make(map[int]*ClientContext)

		for {
			events, err := multiplexer.Wait()
			if err != nil {
				log.Println("Multiplexer Wait error:", err)
				continue
			}

			for _, ev := range events {
				if ev.Fd == serverFd {
					conn, err := listener.Accept()
					if err != nil {
						log.Println("Accept error:", err)
						continue
					}

					tcpConn, ok := conn.(*net.TCPConn)
					if !ok {
						continue
					}
					clientFile, err := tcpConn.File()
					if err != nil {
						continue
					}
					clientFd := int(clientFile.Fd())

					clientCtx := &ClientContext{
						fd:     clientFd,
						conn:   conn,
						parser: resp.NewParser(conn),
					}
					clients[clientFd] = clientCtx

					err = multiplexer.Monitor(io_multiplexing.Event{Fd: clientFd, Op: io_multiplexing.OpRead})
					if err != nil {
						log.Println("Monitor error:", err)
					}
				} else {
					// 4. Dispatch read event to the parser worker pool
					if clientCtx, ok := clients[ev.Fd]; ok {
						connQueue <- ReadEventJob{ctx: clientCtx, fallback: false}
					}
				}
			}
		}
	}
}
