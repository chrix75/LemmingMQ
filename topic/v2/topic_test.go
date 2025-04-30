package v2

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

type CountConsumer[T any] struct {
	count int
}

type wiretapConsumer[T any] struct {
	messages []Message[T]
}

func (w *wiretapConsumer[T]) HandleMessage(m Message[T]) error {
	w.messages = append(w.messages, m)
	return nil
}

func (w *wiretapConsumer[T]) extractMessageIDs() []int {
	ids := make([]int, len(w.messages))
	for i, m := range w.messages {
		ids[i] = m.ID
	}

	return ids
}

func (l *CountConsumer[T]) HandleMessage(Message[T]) error {
	l.count++
	return nil
}

func TestCreateBroadcastTopic(t *testing.T) {
	// given
	tp := NewBroadcastTopic[string]()
	counter := &CountConsumer[string]{}
	wiretap := &wiretapConsumer[string]{}

	tp.AddConsumer(counter)
	tp.AddConsumer(wiretap)

	// when
	var expectedMessages []Message[string]

	for i := 1; i <= 2; i++ {
		msg := fmt.Sprintf("Hello World %d", i)
		tp.SendMessage(msg)
		expectedMessages = append(expectedMessages, Message[string]{
			ID:      i,
			Content: msg,
		})
	}

	// then
	time.Sleep(1 * time.Second)
	assert.Equal(t, expectedMessages, wiretap.messages)
	assert.Equal(t, len(expectedMessages), counter.count)
}

func TestBroadcastToManyConsumers(t *testing.T) {
	// given
	tp := NewBroadcastTopic[string]()
	c1 := &wiretapConsumer[string]{}
	c2 := &wiretapConsumer[string]{}

	tp.AddConsumer(c1)
	tp.AddConsumer(c2)

	// when
	for i := 1; i <= 2; i++ {
		msg := fmt.Sprintf("Hello World %d", i)
		tp.SendMessage(msg)
	}

	// then
	time.Sleep(1 * time.Second)

	actualIDsForC1 := c1.extractMessageIDs()
	actualIDsForC2 := c2.extractMessageIDs()

	expectedIDs := []int{1, 2}
	assert.Equal(t, expectedIDs, actualIDsForC1)
	assert.Equal(t, expectedIDs, actualIDsForC2)
}
