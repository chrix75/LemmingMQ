package v2

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"log"
	"testing"
)

type LogConsumer[T any] struct {
	messages []string
}

func (l *LogConsumer[T]) HandleMessage(message T) error {
	msg := fmt.Sprintf("%v", message)
	l.messages = append(l.messages, msg)
	log.Println(msg)
	return nil
}

func TestCreateBroadcastTopic(t *testing.T) {
	// given
	tp := NewBroadcastTopic[string](10)
	c := &LogConsumer[string]{}

	tp.AddConsumer(c)

	// when
	tp.SendMessage("Hello World")

	// then
	assert.Equal(t, []string{"Hello World"}, c.messages)
}
