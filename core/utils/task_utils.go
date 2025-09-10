package utils

import (
	"fmt"
	"math/rand/v2"
	"sync"
)

type Pool struct {
	taskChan chan func()
	wg       sync.WaitGroup
}

// NewPool 创建一个线程池
func NewPool(workerCount int) *Pool {
	p := &Pool{
		taskChan: make(chan func(), 1000), // 队列缓冲
	}
	// 启动 worker
	for i := 0; i < workerCount; i++ {
		go func() {
			for task := range p.taskChan {
				task()
				p.wg.Done()
			}
		}()
	}
	return p
}

// Submit 提交任务
func (p *Pool) Submit(task func()) {
	p.wg.Add(1)
	p.taskChan <- task
}

// Wait 等待所有任务完成
func (p *Pool) Wait() {
	p.wg.Wait()
}

// Close 关闭线程池
func (p *Pool) Close() {
	close(p.taskChan)
}

// PickAndRemove 数组中随机取一个对象返回并保留剩下的数据
func PickAndRemove[T any](slice []T) ([]T, T, error) {
	if len(slice) == 0 {
		var zero T
		return slice, zero, fmt.Errorf("slice is empty")
	}
	index := rand.IntN(len(slice))
	value := slice[index]
	slice[index] = slice[len(slice)-1]
	slice = slice[:len(slice)-1]
	return slice, value, nil
}
