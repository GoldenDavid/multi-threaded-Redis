package main

import (
	"io"
	"log"
	"net"
	"time"

	"multi-threaded-Redis/Internal/aof"
	"multi-threaded-Redis/Internal/core/io_multiplexing"
	"multi-threaded-Redis/Internal/database"
	"multi-threaded-Redis/Internal/resp"
	"multi-threaded-Redis/ThreadPool"
)

type ClientContext struct {
	fd     int
	conn   net.Conn
	parser *resp.Parser
}

type CommandJob struct {
	client net.Conn
	cmd    resp.Value
}

type ParseTask struct {
	ctx      *ClientContext
	fallback bool
	cmdQueue chan<- CommandJob
}

func (pt *ParseTask) Execute() {
	if pt.fallback {
		// Fallback: block and read continuously
		for {
			cmd, err := pt.ctx.parser.Parse()
			if err != nil {
				if err == io.EOF {
					log.Println("client disconnected:", pt.ctx.conn.RemoteAddr())
				} else {
					log.Println("client read error:", err)
				}
				pt.ctx.conn.Close()
				break
			}
			log.Printf("command received: %+v\n", cmd)
			pt.cmdQueue <- CommandJob{client: pt.ctx.conn, cmd: cmd}
		}
	} else {
		// Event-driven: read available data then return to pool
		for {
			cmd, err := pt.ctx.parser.Parse()
			if err != nil {
				if err == io.EOF {
					log.Println("client disconnected:", pt.ctx.conn.RemoteAddr())
				} else {
					log.Println("client read error:", err)
				}
				pt.ctx.conn.Close()
				break
			}
			log.Printf("command received: %+v\n", cmd)
			pt.cmdQueue <- CommandJob{client: pt.ctx.conn, cmd: cmd}
			
			if !pt.ctx.parser.HasMoreData() {
				break
			}
		}
	}
}

type ResponseTask struct {
	client net.Conn
	result resp.Value
}

func (rt *ResponseTask) Execute() {
	writer := resp.NewWriter(rt.client)
	err := writer.Write(rt.result)
	if err != nil {
		log.Println("err write to client", rt.client.RemoteAddr(), ":", err)
	}
}

func dbExecutor(db *database.Database, cmdQueue <-chan CommandJob, responsePool *threadpool.Pool) {
	log.Println("DB Executor started")
	ticker := time.NewTicker(100 * time.Millisecond) // Run active expiration every 100ms
	defer ticker.Stop()

	for {
		select {
		case job, ok := <-cmdQueue:
			if !ok {
				return
			}
			// Execute command sequentially in a single goroutine
			result := db.Exec(job.cmd)
			
			// Push the result to the Response Pool
			responsePool.AddTask(&ResponseTask{client: job.client, result: result})
		case <-ticker.C:
			// Active expiration in the same goroutine to ensure thread safety
			db.DeleteExpiredKeys(20) // Delete up to 20 expired keys per tick
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
	cmdQueue := make(chan CommandJob, 1000)

	// 1. Start a pool of Response Workers
	responsePool := threadpool.NewPool(4, 1000)

	// 2. Start the single-threaded Database Executor
	go dbExecutor(db, cmdQueue, responsePool)

	// 3. Start a pool of Parser Workers
	parserPool := threadpool.NewPool(4, 1000)

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
			parserPool.AddTask(&ParseTask{ctx: clientCtx, fallback: true, cmdQueue: cmdQueue})
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
						parserPool.AddTask(&ParseTask{ctx: clientCtx, fallback: false, cmdQueue: cmdQueue})
					}
				}
			}
		}
	}
}
