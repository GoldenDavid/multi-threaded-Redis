package main

import (
	"log"
	"net"
)

// element in queue
type Job struct {
	conn net.Conn
}

// thread in pool
type Worker struct {
	id       int
	jobQueue chan Job
}

type Pool struct {
	// queue
	jobQueue chan Job
	workers  []*Worker
}

func (p *Pool) Start() {
	for i := 0; i < len(p.workers); i++ {
		p.workers[i] = NewWorker(i, p.jobQueue)
		p.workers[i].Start()
	}
}

func (p *Pool) AddJob(conn net.Conn) {
	p.jobQueue <- Job{conn: conn}
}

func NewPool(n int) *Pool {
	return &Pool{
		jobQueue: make(chan Job),
		workers:  make([]*Worker, n),
	}
}

func NewWorker(id int, jobQueue chan Job) *Worker {
	return &Worker{
		id:       id,
		jobQueue: jobQueue,
	}
}

func (w *Worker) Start() {
	go func() {
		for job := range w.jobQueue {
			log.Printf("worker %d is processing from %s", w.id, job.conn.RemoteAddr())
			handleConnection(job.conn)
		}
	}()
}

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
	pool := NewPool(2)
	pool.Start()

	for {
		//conn == socket == communication channel
		conn, err := listener.Accept()
		if err != nil {
			log.Println(err)
		}

		log.Println("handle conn from =", conn.RemoteAddr())

		// Add connection to queue
		pool.AddJob(conn)
	}

}
