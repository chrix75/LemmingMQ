package topic

import "fmt"

type TopicError[T any] struct {
	Message Message[T]
	Err     error
}

func (t TopicError[T]) Error() string {
	return fmt.Sprintf("%d/%s", t.Message.ID, t.Err.Error())
}
