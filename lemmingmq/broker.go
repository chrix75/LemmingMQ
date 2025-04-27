package lemmingmq

import (
	"LemmingMQ/topic"
	"context"
	"errors"
	"maps"
)

type BrokerConfiguration struct {
	QueueSize int
}

type Broker struct {
	BrokerConfiguration
	topics map[string]*topic.Topic
	events chan Event
}

func (b *Broker) Topics() []string {
	topics := make([]string, 0, len(b.topics))
	for t := range maps.Values(b.topics) {
		topics = append(topics, t.Name)
	}
	return topics
}

func (b *Broker) AddTopic(cfg topic.Configuration) {
	tp := topic.NewTopic(cfg)
	b.topics[cfg.Name] = tp
}

func (b *Broker) AddCallbackConsumer(topic string, f topic.ConsumerCallback) {
	tp, found := b.topics[topic]
	if found {
		tp.AddConsumer(f)
	}
}

func (b *Broker) ConsumerCount(topic string) int {
	tp, found := b.topics[topic]
	if found {
		return tp.ConsumerCount()
	}
	return 0
}

func (b *Broker) AddHandlerConsumer(topic string, handler topic.MessageHandler) {
	tp, found := b.topics[topic]
	if found {
		tp.AddMessageHandler(handler)
	}
}

func (b *Broker) RemoveHandlerConsumer(topic string, handler topic.MessageHandler) {
	tp, found := b.topics[topic]
	if found {
		tp.RemoveMessageHandler(handler)
	}
}

type Event struct {
	ctx     context.Context
	topic   *topic.Topic
	content []byte
}

func (b *Broker) SendMessage(ctx context.Context, topic string, content []byte) error {
	tp, found := b.topics[topic]
	if !found {
		return errors.New("topic not found")
	}

	event := Event{
		ctx:     ctx,
		topic:   tp,
		content: content,
	}

	b.events <- event

	return nil
}

func (b *Broker) Start() {
	go func() {
		for event := range b.events {
			go func() {
				_ = event.topic.SendMessage(event.ctx, event.content)
			}()
		}
	}()
}

func NewBroker(cfg BrokerConfiguration) *Broker {
	return &Broker{
		BrokerConfiguration: cfg,
		topics:              make(map[string]*topic.Topic),
		events:              make(chan Event, cfg.QueueSize),
	}
}
