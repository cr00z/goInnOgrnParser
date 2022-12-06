package workerpool

import (
	"context"
	"math/rand"
	"sync"
	"time"

	"github.com/cr00z/goInnOgrnParser/internal/options"
)

type Pool struct {
	workerPoolSize int
	interChan      chan Task
	jobsChan       chan Task
	resultChan     chan Result
	resultDone     chan bool
	resultWG       sync.WaitGroup
	workersWG      sync.WaitGroup
}

func MakePool(workerPoolSize int) *Pool {
	return &Pool{
		workerPoolSize: workerPoolSize,
		interChan:      make(chan Task),
		jobsChan:       make(chan Task, workerPoolSize),
		resultChan:     make(chan Result, workerPoolSize),
		resultDone:     make(chan bool, 1),
		resultWG:       sync.WaitGroup{},
		workersWG:      sync.WaitGroup{},
	}
}

func (p *Pool) StartInterChan(ctx context.Context) {
	for {
		select {
		case job := <-p.interChan:
			p.jobsChan <- job
		case <-ctx.Done():
			options.PLog.Println("Pool received cancellation signal, closing job channel")
			close(p.jobsChan)
			options.PLog.Println("Pool closed job channel")
			return
		}
	}
}

func (p *Pool) GetInterChan() chan<- Task {
	return p.interChan
}

func (p *Pool) ProcessResults() {
	p.resultWG.Add(1)
	defer p.resultWG.Done()
	proceedResult := true
	for proceedResult || len(p.resultChan) > 0 {
		select {
		case result := <-p.resultChan:
			options.PLog.Printf("Receive job %d result\n", result.ID)
			result.Process()
		case <-p.resultDone:
			options.PLog.Println("Pool received cancellation signal, closing result channel")
			proceedResult = false
		}
	}
}

func (p *Pool) WaitResults() {
	p.resultWG.Wait()
}

func (p *Pool) RunWorkers() {
	p.workersWG.Add(p.workerPoolSize)
	for i := 1; i <= p.workerPoolSize; i++ {
		go p.WorkerFunc(i)
	}
}

func (p *Pool) WaitWorkers() {
	p.workersWG.Wait()
}

func (p *Pool) WorkerFunc(index int) {
	defer p.workersWG.Done()

	options.PLog.Printf("Worker %d starting\n", index)
	for task := range p.jobsChan {
		options.PLog.Printf("Worker %d started job %d\n", index, task.ID)
		time.Sleep(time.Millisecond * time.Duration(1000+rand.Intn(2000)))
		p.resultChan <- task.Process()
		options.PLog.Printf("Worker %d finished processing job %d\n", index, task.ID)
	}
	options.PLog.Printf("Worker %d interrupted\n", index)
}

func (p *Pool) ResultDone() {
	p.resultDone <- true
}
