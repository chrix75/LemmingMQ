package topic

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestCreateDispatchTopic(t *testing.T) {
	// given
	errChan := make(chan TopicError[string])
	tp := NewDispatchTopic[string](errChan)
	wiretap1 := &wiretapConsumer[string]{}
	wiretap2 := &wiretapConsumer[string]{}

	tp.AddConsumer(wiretap1)
	tp.AddConsumer(wiretap2)

	// when
	var expectedMessages []Message[string]

	for i := 1; i <= 500; i++ {
		msg := fmt.Sprintf("Hello World %d", i)
		tp.SendMessage(msg)
		expectedMessages = append(expectedMessages, Message[string]{
			ID:      i,
			Content: msg,
		})
	}

	// then
	time.Sleep(1 * time.Second)
	allMessages := wiretap1.messages[:]
	allMessages = append(allMessages, wiretap2.messages...)

	assert.NotEmpty(t, wiretap1.messages)
	assert.NotEmpty(t, wiretap2.messages)
	assert.ElementsMatch(t, expectedMessages, allMessages)
}
