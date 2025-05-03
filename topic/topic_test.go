package topic

import (
	"errors"
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
	errChan := make(chan TopicError[string])
	tp := NewBroadcastTopic[string](errChan)
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
	errChan := make(chan TopicError[string])
	tp := NewBroadcastTopic[string](errChan)
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

func TestManageError(t *testing.T) {
	// given
	errChan := make(chan TopicError[string])
	tp := NewBroadcastTopic[string](errChan)
	c1 := &consumerWithError[string]{}

	tp.AddConsumer(c1)

	var encounteredErrors []string
	go func() {
		for err := range errChan {
			encounteredErrors = append(encounteredErrors, err.Error())
		}
	}()

	// when
	for i := 1; i <= 2; i++ {
		msg := fmt.Sprintf("Hello World %d", i)
		tp.SendMessage(msg)
	}

	// then
	time.Sleep(1 * time.Second)
	close(errChan)

	expectedErrors := []string{fmt.Sprintf("1/error")}
	assert.Equal(t, expectedErrors, encounteredErrors)
}

type consumerWithError[T any] struct {
	messageCount int
}

func (c *consumerWithError[T]) HandleMessage(Message[T]) error {
	c.messageCount++
	if c.messageCount == 1 {
		return errors.New("error")
	}
	return nil
}
