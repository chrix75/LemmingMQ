package topic

import (
	"context"
	"errors"
	"slices"
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
	callbacks        []ConsumerCallback
	handlers         []MessageHandler
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
	return len(t.callbacks) + len(t.handlers)
}

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
	t.callbacks = append(t.callbacks, f)
}

// SendMessage sends a message with the specified content to the topic's consumers.
// For BroadcastTopic, the message is sent to allconsumers. For DispatchTopic, the message is sent to a single consumer.
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
//	err := topic.SendMessage(ctx, []byte(`{"message":"Hello World"}`))
//	if err != nil {
//		// Handle error
//	}
func (t *Topic) SendMessage(ctx context.Context, content []byte) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		t.currentMessageID++

		msg := NewMessage(t.currentMessageID, t.Name, content)

		if t.Diffusion == BroadcastTopic {
			return t.sendMessageToAllConsumers(ctx, msg)
		}

		return t.dispatchMessage(ctx, msg)
	}
}

func (t *Topic) sendMessageToAllConsumers(ctx context.Context, msg Message) error {
	var errs []error

	callbackErrs := t.sendMessageToAllCallbacks(ctx, msg)
	errs = append(errs, callbackErrs...)

	handlerErrs := t.sendMessageToAllHandlers(ctx, msg)
	errs = append(errs, handlerErrs...)

	return errors.Join(errs...)
}

func (t *Topic) sendMessageToAllHandlers(ctx context.Context, msg Message) []error {
	var errs []error

	counters := countersFromConsumers(len(t.callbacks), t.Retries)

	for i, consumer := range t.callbacks {
		err := consumer(ctx, msg)

		for err != nil && counters[i] > 0 {
			counters[i]--
			err = consumer(ctx, msg)
		}

		if err != nil {
			errs = append(errs, err)
		}
	}

	return errs
}

func (t *Topic) sendMessageToAllCallbacks(ctx context.Context, msg Message) []error {
	var errs []error

	counters := countersFromConsumers(len(t.handlers), t.Retries)

	for i, handler := range t.handlers {
		err := handler.Handle(ctx, msg)

		for err != nil && counters[i] > 0 {
			counters[i]--
			err = handler.Handle(ctx, msg)
		}

		if err != nil {
			errs = append(errs, err)
		}
	}

	return errs
}

func countersFromConsumers(count int, retries int) []int {
	counters := make([]int, count)
	for i := 0; i < count; i++ {
		counters[i] = retries
	}

	return counters
}

func (t *Topic) dispatchMessage(ctx context.Context, msg Message) error {
	callback, handler := t.selectConsumer()
	if callback == nil && handler == nil {
		return errors.New("could not find a consumer")
	}

	count := t.Retries
	if callback != nil {
		err := callback(ctx, msg)
		for err != nil && count > 0 {
			count--
			err = callback(ctx, msg)
		}

		return err
	}

	err := handler.Handle(ctx, msg)
	for err != nil && count > 0 {
		count--
		err = handler.Handle(ctx, msg)
	}

	return err
}

func (t *Topic) selectConsumer() (ConsumerCallback, MessageHandler) {
	if len(t.callbacks) == 0 && len(t.handlers) == 0 {
		return nil, nil
	}

	if t.dispatcherIndex >= len(t.callbacks)+len(t.handlers) {
		t.dispatcherIndex = 0
	}

	if t.dispatcherIndex < len(t.callbacks) {
		selectedIndex := t.dispatcherIndex
		t.dispatcherIndex++
		return t.callbacks[selectedIndex], nil
	}

	selectedIndex := t.dispatcherIndex - len(t.callbacks)
	return nil, t.handlers[selectedIndex]

}

// AddMessageHandler registers a new message handler to the topic, appending it to the list of existing handlers.
func (t *Topic) AddMessageHandler(handler MessageHandler) {
	t.handlers = append(t.handlers, handler)
}

func (t *Topic) RemoveMessageHandler(handler MessageHandler) {
	handlerIndex := slices.Index(t.handlers, handler)
	if handlerIndex >= 0 {
		t.handlers = append(t.handlers[:handlerIndex], t.handlers[handlerIndex+1:]...)
	}
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
