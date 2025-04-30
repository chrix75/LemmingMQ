package v2

type Message[T any] struct {
	ID      int
	Content T
}
