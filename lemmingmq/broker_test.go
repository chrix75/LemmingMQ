package lemmingmq

import (
	"LemmingMQ/topic"
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
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

	cfg := topic.Configuration{
		Name:      "test",
		Diffusion: topic.BroadcastTopic,
		Retries:   0,
	}

	// when
	broker.AddTopic(cfg)

	// then
	topics := broker.Topics()
	assert.Equal(t, []string{"test"}, topics)
}

func TestAddCallbackConsumer(t *testing.T) {
	// given
	broker := NewBroker(cfg)

	cfg := topic.Configuration{
		Name:      "test",
		Diffusion: topic.BroadcastTopic,
		Retries:   0,
	}

	broker.AddTopic(cfg)

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

	cfg := topic.Configuration{
		Name:      "test",
		Diffusion: topic.BroadcastTopic,
		Retries:   0,
	}

	broker.AddTopic(cfg)

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
