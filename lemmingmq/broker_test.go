package lemmingmq

import (
	"context"
	"errors"
	"github.com/chrix75/LemmingMQ/topic"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
	"time"
)

var cfg = BrokerConfiguration{
	QueueSize: 10,
}

func TestCreateEmptyBroker(t *testing.T) {
	// when
	broker := NewBroker(cfg)

	// then
	topics := broker.Topics()
	assert.Empty(t, topics)
}

func TestCreateBrokerWithOneTopic(t *testing.T) {
	// given
	broker := NewBroker(cfg)

	topicCfg := topic.Configuration{
		Name:      "test",
		Diffusion: topic.BroadcastTopic,
		Retries:   0,
	}

	// when
	broker.AddTopic(topicCfg)

	// then
	topics := broker.Topics()
	assert.Equal(t, []string{"test"}, topics)
}

func TestAddCallbackConsumer(t *testing.T) {
	// given
	broker := NewBroker(cfg)

	topicCfg := topic.Configuration{
		Name:      "test",
		Diffusion: topic.BroadcastTopic,
		Retries:   0,
	}

	broker.AddTopic(topicCfg)

	// when
	broker.AddCallbackConsumer("test", func(ctx context.Context, message topic.Message) error {
		return nil
	})

	// then
	count := broker.ConsumerCount("test")
	assert.Equal(t, 1, count)
}

type noopHandler struct {
}

func (n noopHandler) Handle(context.Context, topic.Message) error {
	return nil
}

func TestAddHandlerConsumer(t *testing.T) {
	// given
	broker := NewBroker(cfg)

	topicCfg := topic.Configuration{
		Name:      "test",
		Diffusion: topic.BroadcastTopic,
		Retries:   0,
	}

	broker.AddTopic(topicCfg)

	handler := noopHandler{}

	// when
	broker.AddHandlerConsumer("test", handler)

	// then
	count := broker.ConsumerCount("test")
	assert.Equal(t, 1, count)
}

func TestRemoveHandlerConsumer(t *testing.T) {
	// given
	broker := NewBroker(cfg)

	cfg := topic.Configuration{
		Name:      "test",
		Diffusion: topic.BroadcastTopic,
		Retries:   0,
	}

	broker.AddTopic(cfg)

	handler := noopHandler{}
	broker.AddHandlerConsumer("test", handler)

	// when
	broker.RemoveHandlerConsumer("test", handler)

	// then
	count := broker.ConsumerCount("test")
	assert.Equal(t, 0, count)
}

func TestFireAndForgetDispatchMessages(t *testing.T) {
	// given
	ctx := context.Background()

	broker := NewBroker(cfg)
	broker.Start()

	topicCfg := topic.Configuration{
		Name:      "test",
		Diffusion: topic.DispatchTopic,
		Retries:   0,
	}

	broker.AddTopic(topicCfg)

	var receivedCount int
	var mu sync.Mutex

	callback := func(ctx context.Context, message topic.Message) error {
		mu.Lock()
		defer mu.Unlock()
		receivedCount++
		time.Sleep(1 * time.Second)
		return nil
	}

	broker.AddCallbackConsumer("test", callback)

	// when
	content := []byte("hello world")

	startTime := time.Now()
	err1 := broker.SendMessage(ctx, "test", content)
	err2 := broker.SendMessage(ctx, "test", content)
	duration := time.Since(startTime)

	err := errors.Join(err1, err2)

	time.Sleep(time.Millisecond * 1500)

	// then
	assert.Nil(t, err)
	assert.True(t, duration < 2*time.Second)
	assert.Equal(t, 2, receivedCount)
}
