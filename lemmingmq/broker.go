package lemmingmq

import (
	"LemmingMQ/topic"
	"context"
	"errors"
	"maps"
)

// BrokerConfiguration holds the configuration parameters for a Broker.
//
// Example:
//
//	// Create a broker configuration with a queue size of 100
//	cfg := BrokerConfiguration{
//		QueueSize: 100,
//	}
type BrokerConfiguration struct {
	QueueSize int
}

// Broker manages multiple topics and handles message routing.
// It provides methods for adding topics, consumers, and sending messages.
//
// Example:
//
//	// Create a new broker with a queue size of 100
//	broker := NewBroker(BrokerConfiguration{QueueSize: 100})
//
//	// Add a topic
//	broker.AddTopic(topic.Configuration{
//		Name:      "notifications",
//		Diffusion: topic.BroadcastTopic,
//		Retries:   3,
//	})
//
//	// Start the broker
//	broker.Start()
type Broker struct {
	BrokerConfiguration
	topics map[string]*topic.Topic
	events chan Event
}

// Topics returns a list of all topic names managed by the broker.
//
// Example:
//
//	// Get all topic names
//	broker := NewBroker(BrokerConfiguration{QueueSize: 100})
//	broker.AddTopic(topic.Configuration{Name: "notifications"})
//	broker.AddTopic(topic.Configuration{Name: "events"})
//
//	topicNames := broker.Topics()
//	// topicNames will contain ["notifications", "events"]
func (b *Broker) Topics() []string {
	topics := make([]string, 0, len(b.topics))
	for t := range maps.Values(b.topics) {
		topics = append(topics, t.Name)
	}
	return topics
}

// AddTopic creates a new topic with the given configuration and adds it to the broker.
//
// Example:
//
//	// Create a broker
//	broker := NewBroker(BrokerConfiguration{QueueSize: 100})
//
//	// Add a broadcast topic
//	broker.AddTopic(topic.Configuration{
//		Name:      "notifications",
//		Diffusion: topic.BroadcastTopic,
//		Retries:   3,
//	})
//
//	// Add a dispatch topic
//	broker.AddTopic(topic.Configuration{
//		Name:      "tasks",
//		Diffusion: topic.DispatchTopic,
//		Retries:   1,
//	})
func (b *Broker) AddTopic(cfg topic.Configuration) {
	tp := topic.NewTopic(cfg)
	b.topics[cfg.Name] = tp
}

// AddCallbackConsumer adds a callback function as a consumer to the specified topic.
// If the topic doesn't exist, the consumer is not added.
//
// Example:
//
//	// Create a broker with a topic
//	broker := NewBroker(BrokerConfiguration{QueueSize: 100})
//	broker.AddTopic(topic.Configuration{Name: "notifications"})
//
//	// Add a callback consumer
//	broker.AddCallbackConsumer("notifications", func(ctx context.Context, msg topic.Message) error {
//		fmt.Printf("Received message: %s\n", string(msg.Content))
//		return nil
//	})
func (b *Broker) AddCallbackConsumer(topic string, f topic.ConsumerCallback) {
	tp, found := b.topics[topic]
	if found {
		tp.AddConsumer(f)
	}
}

// ConsumerCount returns the number of consumers for the specified topic.
// If the topic doesn't exist, it returns 0.
//
// Example:
//
//	// Create a broker with a topic
//	broker := NewBroker(BrokerConfiguration{QueueSize: 100})
//	broker.AddTopic(topic.Configuration{Name: "notifications"})
//
//	// Add a consumer
//	broker.AddCallbackConsumer("notifications", func(ctx context.Context, msg topic.Message) error {
//		return nil
//	})
//
//	// Get the consumer count
//	count := broker.ConsumerCount("notifications") // count will be 1
//
//	// Get count for non-existent topic
//	count = broker.ConsumerCount("unknown") // count will be 0
func (b *Broker) ConsumerCount(topic string) int {
	tp, found := b.topics[topic]
	if found {
		return tp.ConsumerCount()
	}
	return 0
}

// AddHandlerConsumer adds a message handler as a consumer to the specified topic.
// If the topic doesn't exist, the handler is not added.
//
// Example:
//
//	// Create a broker with a topic
//	broker := NewBroker(BrokerConfiguration{QueueSize: 100})
//	broker.AddTopic(topic.Configuration{Name: "notifications"})
//
//	// Create a message handler
//	type LogHandler struct{}
//
//	func (h LogHandler) Handle(ctx context.Context, msg topic.Message) error {
//		fmt.Printf("Received message: %s\n", string(msg.Content))
//		return nil
//	}
//
//	// Add the handler as a consumer
//	handler := LogHandler{}
//	broker.AddHandlerConsumer("notifications", handler)
func (b *Broker) AddHandlerConsumer(topic string, handler topic.MessageHandler) {
	tp, found := b.topics[topic]
	if found {
		tp.AddMessageHandler(handler)
	}
}

// RemoveHandlerConsumer removes a message handler from the specified topic.
// If the topic doesn't exist, no action is taken.
//
// Example:
//
//	// Create a broker with a topic
//	broker := NewBroker(BrokerConfiguration{QueueSize: 100})
//	broker.AddTopic(topic.Configuration{Name: "notifications"})
//
//	// Create and add a message handler
//	handler := LogHandler{}
//	broker.AddHandlerConsumer("notifications", handler)
//
//	// Later, remove the handler
//	broker.RemoveHandlerConsumer("notifications", handler)
func (b *Broker) RemoveHandlerConsumer(topic string, handler topic.MessageHandler) {
	tp, found := b.topics[topic]
	if found {
		tp.RemoveMessageHandler(handler)
	}
}

// Event represents a message event that will be processed by the broker.
// It contains the context, target topic, and message content.
//
// Example:
//
//	// Events are typically created internally by the SendMessage method
//	event := Event{
//		ctx:     context.Background(),
//		topic:   topicInstance,
//		content: []byte("Hello, world!"),
//	}
type Event struct {
	ctx     context.Context
	topic   *topic.Topic
	content []byte
}

// SendMessage sends a message to the specified topic.
// It returns an error if the topic doesn't exist.
//
// Example:
//
//	// Create a broker with a topic
//	broker := NewBroker(BrokerConfiguration{QueueSize: 100})
//	broker.AddTopic(topic.Configuration{Name: "notifications"})
//	broker.Start()
//
//	// Send a message to the topic
//	ctx := context.Background()
//	err := broker.SendMessage(ctx, "notifications", []byte("Hello, world!"))
//	if err != nil {
//		fmt.Printf("Error sending message: %v\n", err)
//	}
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

// Start begins processing messages in the broker.
// It launches a goroutine that listens for events and dispatches them to the appropriate topics.
//
// Example:
//
//	// Create a broker with a topic and consumer
//	broker := NewBroker(BrokerConfiguration{QueueSize: 100})
//	broker.AddTopic(topic.Configuration{Name: "notifications"})
//	broker.AddCallbackConsumer("notifications", func(ctx context.Context, msg topic.Message) error {
//		fmt.Printf("Received message: %s\n", string(msg.Content))
//		return nil
//	})
//
//	// Start the broker to begin processing messages
//	broker.Start()
//
//	// Now messages can be sent
//	broker.SendMessage(context.Background(), "notifications", []byte("Hello, world!"))
func (b *Broker) Start() {
	go func() {
		for event := range b.events {
			go func() {
				_ = event.topic.SendMessage(event.ctx, event.content)
			}()
		}
	}()
}

// NewBroker creates a new broker with the given configuration.
//
// Example:
//
//	// Create a broker with a queue size of 100
//	broker := NewBroker(BrokerConfiguration{
//		QueueSize: 100,
//	})
//
//	// Add topics and start the broker
//	broker.AddTopic(topic.Configuration{Name: "notifications"})
//	broker.Start()
func NewBroker(cfg BrokerConfiguration) *Broker {
	return &Broker{
		BrokerConfiguration: cfg,
		topics:              make(map[string]*topic.Topic),
		events:              make(chan Event, cfg.QueueSize),
	}
}
