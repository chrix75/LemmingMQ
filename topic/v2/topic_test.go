package v2

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"log"
	"testing"
	"time"
)

type LogConsumer[T any] struct {
	messages []string
}

func (l *LogConsumer[T]) HandleMessage(message Message[T]) error {
	msg := fmt.Sprintf("%v", message)
	l.messages = append(l.messages, msg)
	log.Println(msg)
	return nil
}

func TestCreateBroadcastTopic(t *testing.T) {
	// given
	tp := NewBroadcastTopic[string]()
	c := &LogConsumer[string]{}

	tp.AddConsumer(c)

	// when
	sentMessages := []string{}
	for i := 1; i < 10; i++ {
		msg := fmt.Sprintf("Hello World %d", i)
		tp.SendMessage(msg)
		sentMessages = append(sentMessages, msg)
	}

	// then
	time.Sleep(1 * time.Second)
	assert.Equal(t, sentMessages, c.messages)
}

func TestBroadcastToManyConsumers(t *testing.T) {
	// given
	tp := NewBroadcastTopic[string]()
	c1 := &LogConsumer[string]{}
	c2 := &LogConsumer[string]{}

	tp.AddConsumer(c1)
	tp.AddConsumer(c2)

	// when
	sentMessages := []string{}
	for i := 1; i < 10; i++ {
		msg := fmt.Sprintf("Hello World %d", i)
		tp.SendMessage(msg)
		sentMessages = append(sentMessages, msg)
	}

	// then
	time.Sleep(1 * time.Second)
	assert.Equal(t, sentMessages, c1.messages)
	assert.Equal(t, sentMessages, c2.messages)
}
