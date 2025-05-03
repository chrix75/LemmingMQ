package topic

type Message[T any] struct {
	ID      int
	Content T
}
