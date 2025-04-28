package v2

type BroadcastTopic[T any] struct {
	ch        <-chan T
	pool      WorkerPool[T]
	consumers []Consumer[T]
}

type WorkerPool[T any] struct {
	ch   chan T
	size int
}

type Consumer[T any] interface {
	HandleMessage(Message[T]) error
}

type Message[T any] struct {
	Content T
}
