package topic

import (
	"sync"
)

type DispatchTopic[T any] struct {
	ch        chan Message[T]
	messageID int
	mu        sync.Mutex
	errChan   chan TopicError[T]
}

func (t *DispatchTopic[T]) AddConsumer(consumer Consumer[T]) {
	t.mu.Lock()
	if t.ch == nil {
		t.ch = make(chan Message[T])
	}
	t.mu.Unlock()

	go func(consumer Consumer[T]) {
		for msg := range t.ch {
			err := consumer.HandleMessage(msg)
			if err != nil {
				t.errChan <- TopicError[T]{
					Message: msg,
					Err:     err,
				}
			}
		}
	}(consumer)
}

func (t *DispatchTopic[T]) SendMessage(content T) {
	msg := Message[T]{
		ID:      t.nextMessageID(),
		Content: content,
	}

	t.ch <- msg
}

func (t *DispatchTopic[T]) nextMessageID() int {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.messageID++
	return t.messageID
}

func NewDispatchTopic[T any](errChan chan TopicError[T]) *DispatchTopic[T] {
	return &DispatchTopic[T]{
		errChan: errChan,
	}
}
