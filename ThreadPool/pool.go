package threadpool

type Task interface {
	Execute()
}

type Worker struct {
	id        int
	taskQueue chan Task
	quit      chan bool
}

func NewWorker(id int, taskQueue chan Task) *Worker {
	return &Worker{
		id:        id,
		taskQueue: taskQueue,
		quit:      make(chan bool),
	}
}

func (w *Worker) Start() {
	go func() {
		for {
			select {
			case task := <-w.taskQueue:
				if task != nil {
					task.Execute()
				}
			case <-w.quit:
				return
			}
		}
	}()
}

func (w *Worker) Stop() {
	go func() {
		w.quit <- true
	}()
}

type Pool struct {
	taskQueue chan Task
	workers   []*Worker
}

func NewPool(numWorkers int, queueSize int) *Pool {
	pool := &Pool{
		taskQueue: make(chan Task, queueSize),
		workers:   make([]*Worker, numWorkers),
	}
	for i := 0; i < numWorkers; i++ {
		pool.workers[i] = NewWorker(i, pool.taskQueue)
		pool.workers[i].Start()
	}
	return pool
}

func (p *Pool) AddTask(task Task) {
	p.taskQueue <- task
}

func (p *Pool) Stop() {
	for _, worker := range p.workers {
		worker.Stop()
	}
}
