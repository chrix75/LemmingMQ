package topic

import (
	"context"
	"errors"
)

type DiffusionType int8

const (
	BroadcastTopic DiffusionType = iota + 1
	DispatchTopic
)

type Configuration struct {
	Name      string
	Diffusion DiffusionType
	Retries   int
}

type Topic struct {
	Configuration
	consumers        []ConsumerCallback
	dispatcherIndex  int
	currentMessageID int
}

// ConsumerCount returns the number of consumers registered to the topic.
//
// Example:
//
//	topic := NewTopic(Configuration{Name: "example", Diffusion: BroadcastTopic})
//	count := topic.ConsumerCount() // Returns 0 as no consumers are added yet
func (t *Topic) ConsumerCount() int {
	return len(t.consumers)
}

type ConsumerCallback func(c context.Context, msg Message) error

// AddConsumer registers a new consumer callback function to the topic.
//
// Example:
//
//	topic := NewTopic(Configuration{Name: "example", Diffusion: BroadcastTopic})
//	topic.AddConsumer(func(ctx context.Context, msg Message) error {
//		fmt.Println("Received message:", string(msg.Content))
//		return nil
//	})
func (t *Topic) AddConsumer(f ConsumerCallback) {
	t.consumers = append(t.consumers, f)
}

// SendMessage sends a message with the specified content type and content to the topic's consumers.
// For BroadcastTopic, the message is sent to all consumers. For DispatchTopic, the message is sent to a single consumer.
// It respects context cancellation and returns an error if the context is done or if any consumer fails to process the message.
//
// Example:
//
//	ctx := context.Background()
//	topic := NewTopic(Configuration{
//		Name:      "notifications",
//		Diffusion: BroadcastTopic,
//		Retries:   3,
//	})
//	
//	// Add consumers before sending messages
//	topic.AddConsumer(func(ctx context.Context, msg Message) error {
//		// Process the message
//		return nil
//	})
//	
//	// Send a JSON message
//	err := topic.SendMessage(ctx, "application/json", []byte(`{"message":"Hello World"}`))
//	if err != nil {
//		// Handle error
//	}
func (t *Topic) SendMessage(ctx context.Context, contentType string, content []byte) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		t.currentMessageID++

		msg := NewMessage(t.currentMessageID, contentType, content)

		if t.Diffusion == BroadcastTopic {
			return t.sendMessageToAllConsumers(ctx, msg)
		}

		return t.dispatchMessage(ctx, msg)
	}
}

func (t *Topic) sendMessageToAllConsumers(ctx context.Context, msg Message) error {
	var errs []error

	counters := countersFromConsumers(len(t.consumers), t.Retries)

	for i, consumer := range t.consumers {
		err := consumer(ctx, msg)

		for err != nil && counters[i] > 0 {
			counters[i]--
			err = consumer(ctx, msg)
		}

		if err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

func countersFromConsumers(count int, retries int) []int {
	counters := make([]int, count)
	for i := 0; i < count; i++ {
		counters[i] = retries
	}

	return counters
}

func (t *Topic) dispatchMessage(ctx context.Context, msg Message) error {
	consumer := t.selectConsumer()
	if consumer == nil {
		return errors.New("could not find a consumer")
	}

	count := t.Retries
	err := consumer(ctx, msg)
	for err != nil && count > 0 {
		count--
		err = consumer(ctx, msg)
	}

	return err
}

func (t *Topic) selectConsumer() ConsumerCallback {
	if len(t.consumers) == 0 {
		return nil
	}

	if t.dispatcherIndex >= len(t.consumers) {
		t.dispatcherIndex = 0
	}

	selected := t.consumers[t.dispatcherIndex]
	t.dispatcherIndex++

	return selected
}

// NewTopic creates a new Topic with the provided configuration.
// The configuration includes the topic name, diffusion type (broadcast or dispatch), and retry count.
//
// Example:
//
//	// Create a broadcast topic with 3 retries
//	broadcastTopic := NewTopic(Configuration{
//		Name:      "notifications",
//		Diffusion: BroadcastTopic,
//		Retries:   3,
//	})
//
//	// Create a dispatch topic with 2 retries
//	dispatchTopic := NewTopic(Configuration{
//		Name:      "tasks",
//		Diffusion: DispatchTopic,
//		Retries:   2,
//	})
func NewTopic(cfg Configuration) *Topic {
	return &Topic{
		Configuration: cfg,
	}
}
