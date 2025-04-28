package v2

type BroadcastTopic[T any] struct {
	ch   chan T
	pool *WorkerPool[T]
}

func (t *BroadcastTopic[T]) AddConsumer(c Consumer[T]) {
}

func (t *BroadcastTopic[T]) SendMessage(message T) {
	t.ch <- message
}

type WorkerPool[T any] struct {
	consumers []Consumer[T]
	size      int
	running   bool
}

type Consumer[T any] interface {
	HandleMessage(T) error
}

func NewBroadcastTopic[T any](size int) *BroadcastTopic[T] {
	return &BroadcastTopic[T]{
		ch:   make(chan T),
		pool: NewWorkerPool[T](),
	}
}

func NewWorkerPool[T any]() *WorkerPool[T] {
	return &WorkerPool[T]{
		consumers: make([]Consumer[T], 0),
	}
}
