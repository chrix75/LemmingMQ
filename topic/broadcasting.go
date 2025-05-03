package topic

import (
	"sync"
)

type BroadcastTopic[T any] struct {
	pool      *WorkerPool[T]
	messageID int
	mu        sync.Mutex
	errChan   chan TopicError[T]
}

func (t *BroadcastTopic[T]) AddConsumer(c Consumer[T]) {
	consumerCh := make(chan Message[T])
	t.pool.startConsumer(consumerCh, c)
}

func (t *BroadcastTopic[T]) SendMessage(content T) {
	message := Message[T]{
		ID:      t.nextMessageID(),
		Content: content,
	}
	t.pool.sendMessage(message)
}

func (t *BroadcastTopic[T]) nextMessageID() int {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.messageID++
	return t.messageID
}

type WorkerPool[T any] struct {
	consumerChannels []chan Message[T]
	mu               sync.Mutex
	errCh            chan TopicError[T]
}

func (p *WorkerPool[T]) sendMessage(message Message[T]) {
	p.mu.Lock()
	defer p.mu.Unlock()
	for _, ch := range p.consumerChannels {
		ch <- message
	}
}

func (p *WorkerPool[T]) startConsumer(ch chan Message[T], consumer Consumer[T]) {
	p.consumerChannels = append(p.consumerChannels, ch)
	go func() {
		for message := range ch {
			err := consumer.HandleMessage(message)
			if err != nil {
				p.errCh <- TopicError[T]{
					Message: message,
					Err:     err,
				}
			}
		}
	}()
}

type Consumer[T any] interface {
	HandleMessage(Message[T]) error
}

func NewBroadcastTopic[T any](errChan chan TopicError[T]) *BroadcastTopic[T] {
	return &BroadcastTopic[T]{
		pool: NewWorkerPool[T](errChan),
	}
}

func NewWorkerPool[T any](errCh chan TopicError[T]) *WorkerPool[T] {
	return &WorkerPool[T]{
		errCh:            errCh,
		consumerChannels: make([]chan Message[T], 0),
	}
}
