package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"sync"
	"sync/atomic"
)

type job func() uint64

// Пулл воркеров
// size кол-во воркеров
// jobs основная очередь задач на выполнение
// workerJobs очередь задач , отслеживаемая воркерами
// total кол -во найденных слов
type Pool struct {
	wg         *sync.WaitGroup
	size       int
	jobs       chan job
	workerJobs chan job
	total      uint64
}

func NewPool(size int) *Pool {
	return &Pool{&sync.WaitGroup{},
		size,
		make(chan job),
		make(chan job),
		0,
	}
}
// Метод Start вычитывает канал с задачами и инциирует работу воркеров
// Число одновременно запущенных воркеров контролируется атрибутом size
func (p *Pool) Start() {
	go func() {
		currSize := 0
		for j := range p.jobs {
			if currSize < p.size {
				currSize++
				p.wg.Add(1)
				w := Worker{p}
				go func(j job) {
					defer p.wg.Done()
					w.Work(j)
				}(j)
			} else {
				p.workerJobs <- j
			}
		}
		close(p.workerJobs)
	}()
}
// Добавление задачи в очердь
func (p *Pool) Exec(j job) {
	p.jobs <- j
}

func (p *Pool) Close() {
	close(p.jobs)
}
// Синхронизация запущенных воркеров
func (p *Pool) Wait() {
	p.wg.Wait()
	fmt.Printf("Total: %d\n", p.total)
}

// Worker ссылается на атрибуты workerJobs и total
type Worker struct {
	pool *Pool
}
// Метод Work выполняет переданную ему задачу
// и начинает отслеживание канала w.pool.workerJobs для новых задач
func (w *Worker) Work(j job) {
	count := j()
	atomic.AddUint64(&w.pool.total, count)

	for j := range w.pool.workerJobs {
		count := j()
		atomic.AddUint64(&w.pool.total, count)
	}
}
// Функция поиска слова в переданно ресурсе
func search(source string) uint64 {
	_, err := url.ParseRequestURI(source)
	// file
	if err != nil {
		f, err := os.Open(source)
		if err != nil {
			return 0
		}
		b, err := ioutil.ReadAll(f)
		if err != nil {
			return 0
		}
		return uint64(bytes.Count(b, []byte("Go")))
	}
	// url
	res, err := http.Get(source)
	if err != nil {
		return 0
	}
	defer res.Body.Close()
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return 0
	}
	return uint64(bytes.Count(b, []byte("Go")))

}
// декоратор для функции поиска
func decor(f func(string) uint64, arg string) job {
	return func() uint64 {
		val := f(arg)
		fmt.Printf("Count for %s: %d\n", arg, val)
		return val
	}
}

func main() {
	pool := NewPool(5)

	pool.Start()

	for {
		var source string
		fmt.Scan(&source)
		if source == "" {
			break
		}
		j := decor(search, source)
		pool.Exec(j)
	}
	pool.Close()

	pool.Wait()
}
