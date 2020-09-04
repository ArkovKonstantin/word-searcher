package main

import (
	"fmt"
	"sync/atomic"
	"testing"
	"time"
)

var (
	processedSources = []string{}
	numGoroutine uint64 = 0
)
// По завершению работы по обработке источника кладем его в список обработанных источников processedSources
func testDecor(f func(string) uint64, arg string) job {
	return func() uint64 {
		time.Sleep(10*time.Millisecond)
		defer func() {
			processedSources = append(processedSources, arg)
		}()
			val := f(arg)
		fmt.Printf("Count for %s: %d\n", arg, val)
		return val
	}
}
// В месте создания горутин добавлен счетчик numGoroutine
func (p *Pool) testStart() {
	go func() {
		currSize := 0
		for j := range p.jobs {
			if currSize < p.size {
				currSize++
				p.wg.Add(1)
				w := Worker{p}
				go func(j job) {
					atomic.AddUint64(&numGoroutine, 1)
					defer p.wg.Done()
					w.Work(j)
				}(j)
			} else {
				p.workerJobs <- j
			}
		}
		fmt.Println(currSize)
		close(p.workerJobs)
	}()
}

// Тест на то, что каждый источник будет обработан воркером
func TestSourceProcessing(t *testing.T) {
	processedSources = processedSources[:0]
	numGoroutine = 0
	sources := []string{"/etc/issue", "/proc/cpuinfo", "/etc/passwd", "/proc/meminfo"}
	size := 1
	pool := NewPool(1)

	pool.testStart()
	for _, s := range sources{
		j := testDecor(search, s)
		pool.Exec(j)
	}
	pool.Close()

	// Если закомментрировать pool.Wait, то можно увидеть как тест упадет.
	// В этом случае программа завершится раньше, чем воркер успеет обработать все источники
	pool.Wait()

	if len(processedSources) != len(sources){
		t.Errorf("number of processed sources %d, must be %d\n", len(processedSources), len(sources))
	}
	// Проверка на то, что максимальное кол-во созданных горутин не превышает размера пулла воркеров (Pool.size)
	if int(numGoroutine) != size {
		t.Errorf("num of goroutine %d, must be %d",numGoroutine, len(sources))
	}
}

// Проверка на то, что не не создаются лишние горутины при обработке источников
// В данном тесте должно запустится 4 горутины на каждый источник вместо 10-ти
func TestNumGoroutine(t *testing.T) {
	processedSources = processedSources[:0]
	numGoroutine = 0
	sources := []string{"/etc/issue", "/proc/cpuinfo", "/etc/passwd", "/proc/meminfo"}

	pool := NewPool(10)

	pool.testStart()
	for _, s := range sources{
		j := testDecor(search, s)
		pool.Exec(j)
	}
	pool.Close()

	pool.Wait()

	if len(processedSources) != len(sources){
		t.Errorf("number of processed sources %d, must be %d\n", len(processedSources), len(sources))
	}

	if int(numGoroutine) != len(sources) {
		t.Errorf("num of goroutine %d, must be %d",numGoroutine, len(sources))
	}

}