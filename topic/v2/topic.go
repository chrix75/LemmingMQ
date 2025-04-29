package v2

import "sync"

type BroadcastTopic[T any] struct {
	pool *WorkerPool[T]
}

func (t *BroadcastTopic[T]) AddConsumer(c Consumer[T]) {
	consumerCh := make(chan T)
	t.pool.startConsumer(consumerCh, c)
}

func (t *BroadcastTopic[T]) SendMessage(message T) {
	t.pool.sendMessage(message)
}

type WorkerPool[T any] struct {
	consumerChannels []chan T
	mu               sync.Mutex
}

func (p *WorkerPool[T]) sendMessage(message T) {
	p.mu.Lock()
	defer p.mu.Unlock()
	for _, ch := range p.consumerChannels {
		ch <- message
	}
}

func (p *WorkerPool[T]) startConsumer(ch chan T, consumer Consumer[T]) {
	p.consumerChannels = append(p.consumerChannels, ch)
	go func() {
		for message := range ch {
			_ = consumer.HandleMessage(message)
		}
	}()
}

type Consumer[T any] interface {
	HandleMessage(T) error
}

func NewBroadcastTopic[T any]() *BroadcastTopic[T] {
	return &BroadcastTopic[T]{
		pool: NewWorkerPool[T](),
	}
}

func NewWorkerPool[T any]() *WorkerPool[T] {
	return &WorkerPool[T]{
		consumerChannels: make([]chan T, 0),
	}
}
