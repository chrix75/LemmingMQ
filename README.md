# LemmingMQ

LemmingMQ is a lightweight, in-memory message queue library for Go applications. It provides a simple and efficient way to implement pub/sub patterns and message dispatching within your Go applications.

## Features

- **Two Diffusion Types**:
  - **Broadcast**: Sends messages to all registered consumers
  - **Dispatch**: Distributes messages to consumers in a round-robin fashion
- **Retry Mechanism**: Configurable retry count for failed message deliveries
- **Context Support**: Respects context cancellation for graceful shutdowns
- **Minimal Dependencies**: Only depends on the Go standard library for core functionality
- **Type-Safe API**: Strongly typed interfaces for reliable message handling

## Installation

```bash
go get github.com/chrix75/LemmingMQ
```

## Usage

### Creating a Topic

```go
package main

import (
    "context"
    "fmt"
    "github.com/chrix75/LemmingMQ/topic"
)

func exampleCreateTopics() {
    // Create a broadcast topic (sends messages to all consumers)
    broadcastTopic := topic.NewTopic(topic.Configuration{
        Name:      "notifications",
        Diffusion: topic.BroadcastTopic,
        Retries:   3,
    })

    // Create a dispatch topic (sends messages to one consumer at a time)
    dispatchTopic := topic.NewTopic(topic.Configuration{
        Name:      "tasks",
        Diffusion: topic.DispatchTopic,
        Retries:   2,
    })

}
```

### Adding Consumers

```go
package main

import (
    "context"
    "fmt"
    "github.com/chrix75/LemmingMQ/topic"
)

func exampleAddConsumer() {
    // Create a topic
    myTopic := topic.NewTopic(topic.Configuration{
        Name:      "example",
        Diffusion: topic.BroadcastTopic,
        Retries:   1,
    })

    // Add a consumer to process messages
    myTopic.AddConsumer(func(ctx context.Context, msg topic.Message) error {
        fmt.Printf("Received message: %s\n", string(msg.Content))
        // Process the message
        return nil
    })
}
```

### Sending Messages

```go
package main

import (
    "context"
    "fmt"
    "github.com/chrix75/LemmingMQ/topic"
)

func exampleSendMessages() {
    // Create a topic
    myTopic := topic.NewTopic(topic.Configuration{
        Name:      "example",
        Diffusion: topic.BroadcastTopic,
        Retries:   1,
    })

    // Create a context
    ctx := context.Background()

    // Send a text message
    err := myTopic.SendMessage(ctx, "text/plain", []byte("Hello, world!"))
    if err != nil {
        fmt.Println("Error sending message:", err)
    }

    // Send a JSON message
    err = myTopic.SendMessage(ctx, "application/json", []byte(`{"greeting":"Hello, world!"}`))
    if err != nil {
        fmt.Println("Error sending message:", err)
    }
}
```

### Handling Errors and Retries

LemmingMQ will automatically retry failed message deliveries based on the configured retry count:

```go
package main

import (
    "context"
    "errors"
    "fmt"
    "math/rand"
    "github.com/chrix75/LemmingMQ/topic"
)

func exampleErrorHandling() {
    // Create a topic with 2 retries
    myTopic := topic.NewTopic(topic.Configuration{
        Name:      "example",
        Diffusion: topic.BroadcastTopic,
        Retries:   2,
    })

    // Consumer that sometimes fails
    myTopic.AddConsumer(func(ctx context.Context, msg topic.Message) error {
        // Simulate occasional failures
        if rand.Intn(10) < 3 {
            return errors.New("processing failed")
        }

        fmt.Printf("Processed message: %s\n", string(msg.Content))
        return nil
    })

    // With retries=2, the message will be attempted up to 3 times total
    // (initial attempt + 2 retries)
    ctx := context.Background()
    err := myTopic.SendMessage(ctx, "text/plain", []byte("Hello, world!"))
    if err != nil {
        fmt.Println("Failed after all retries:", err)
    }
}
```

### Context Cancellation

```go
package main

import (
    "context"
    "errors"
    "fmt"
    "time"
    "github.com/chrix75/LemmingMQ/topic"
)

func exampleContextCancellation() {
    // Create a topic
    myTopic := topic.NewTopic(topic.Configuration{
        Name:      "example",
        Diffusion: topic.BroadcastTopic,
        Retries:   1,
    })

    // Create a context with timeout
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    // If context is cancelled before or during message processing,
    // SendMessage will return the context error
    err := myTopic.SendMessage(ctx, "text/plain", []byte("Hello, world!"))
    if err != nil {
        // Check for context cancellation
        if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
            fmt.Println("Message sending was cancelled")
        }
    }

    // Example of manually cancelling the context
    ctx2, cancel2 := context.WithCancel(context.Background())
    // Cancel immediately for demonstration
    cancel2()

    // This will fail with context.Canceled error
    err = myTopic.SendMessage(ctx2, "text/plain", []byte("This won't be sent"))
    if err != nil {
        fmt.Println("Error:", err)
    }
}
```

## Broker

The Broker is a central component that manages multiple topics and provides a unified interface for sending messages and managing consumers.

### Features

- **Topic Management**: Create and manage multiple topics with different configurations
- **Consumer Management**: Add and remove consumers to/from topics
- **Asynchronous Message Processing**: Process messages in separate goroutines
- **Configurable Queue Size**: Set the size of the internal message queue

### Usage

#### Creating a Broker

```go
package main

import (
    "LemmingMQ/lemmingmq"
)

func exampleCreateBroker() {
    // Create a broker with a queue size of 100
    broker := lemmingmq.NewBroker(lemmingmq.BrokerConfiguration{
        QueueSize: 100,
    })

    // Start the broker to begin processing messages
    broker.Start()
}
```

#### Adding Topics to a Broker

```go
package main

import (
    "LemmingMQ/lemmingmq"
    "LemmingMQ/topic"
)

func exampleAddTopicsToBroker() {
    // Create a broker
    broker := lemmingmq.NewBroker(lemmingmq.BrokerConfiguration{
        QueueSize: 100,
    })

    // Add a broadcast topic
    broker.AddTopic(topic.Configuration{
        Name:      "notifications",
        Diffusion: topic.BroadcastTopic,
        Retries:   3,
    })

    // Add a dispatch topic
    broker.AddTopic(topic.Configuration{
        Name:      "tasks",
        Diffusion: topic.DispatchTopic,
        Retries:   2,
    })

    // Get a list of all topic names
    topicNames := broker.Topics() // Returns ["notifications", "tasks"]
}
```

#### Adding Consumers to Topics

```go
package main

import (
    "context"
    "fmt"
    "github.com/chrix75/LemmingMQ/lemmingmq"
    "github.com/chrix75/LemmingMQ/topic"
)

func exampleAddConsumersToBroker() {
    // Create a broker with a topic
    broker := lemmingmq.NewBroker(lemmingmq.BrokerConfiguration{
        QueueSize: 100,
    })

    broker.AddTopic(topic.Configuration{
        Name:      "notifications",
        Diffusion: topic.BroadcastTopic,
        Retries:   3,
    })

    // Add a callback consumer
    broker.AddCallbackConsumer("notifications", func(ctx context.Context, msg topic.Message) error {
        fmt.Printf("Received message: %s\n", string(msg.Content))
        return nil
    })

    // Create a message handler
    type LogHandler struct{}

    func (h LogHandler) Handle(ctx context.Context, msg topic.Message) error {
        fmt.Printf("Handler received message: %s\n", string(msg.Content))
        return nil
    }

    // Add the handler as a consumer
    handler := LogHandler{}
    broker.AddHandlerConsumer("notifications", handler)

    // Get the number of consumers for a topic
    count := broker.ConsumerCount("notifications") // Returns 2

    // Later, remove a handler if needed
    broker.RemoveHandlerConsumer("notifications", handler)
}
```

#### Sending Messages Through the Broker

```go
package main

import (
    "context"
    "fmt"
    "github.com/chrix75/LemmingMQ/lemmingmq"
    "github.com/chrix75/LemmingMQ/topic"
)

func exampleSendMessagesThroughBroker() {
    // Create a broker with a topic and consumer
    broker := lemmingmq.NewBroker(lemmingmq.BrokerConfiguration{
        QueueSize: 100,
    })

    broker.AddTopic(topic.Configuration{
        Name:      "notifications",
        Diffusion: topic.BroadcastTopic,
        Retries:   3,
    })

    broker.AddCallbackConsumer("notifications", func(ctx context.Context, msg topic.Message) error {
        fmt.Printf("Received message: %s\n", string(msg.Content))
        return nil
    })

    // Start the broker to begin processing messages
    broker.Start()

    // Send a message to the topic
    ctx := context.Background()
    err := broker.SendMessage(ctx, "notifications", []byte("Hello, world!"))
    if err != nil {
        fmt.Printf("Error sending message: %v\n", err)
    }
}
```

## API Reference

### Types

#### `topic.DiffusionType`

Enum defining how messages are distributed to consumers:
- `topic.BroadcastTopic`: Sends messages to all consumers
- `topic.DispatchTopic`: Sends messages to one consumer at a time (round-robin)

#### `topic.Configuration`

Configuration for creating a new topic:
- `Name`: String identifier for the topic
- `Diffusion`: The diffusion type (broadcast or dispatch)
- `Retries`: Number of retry attempts for failed message deliveries

#### `topic.Message`

Structure representing a message:
- `ID`: Unique identifier for the message
- `Topic`: The name of the topic 
- `Content`: Byte slice containing the message data

### Functions

#### `topic.NewTopic(cfg Configuration) *Topic`

Creates a new topic with the specified configuration.

#### `topic.NewMessage(id int, topic string, content []byte) Message`

Creates a new message with the specified ID, topic name, and content.

### Methods

#### `(t *Topic) AddConsumer(f ConsumerCallback)`

Registers a new consumer callback function to the topic.

#### `(t *Topic) SendMessage(ctx context.Context, content []byte) error`

Sends a message with the specified content to the topic's consumers.

#### `(t *Topic) ConsumerCount() int`

Returns the number of consumers registered to the topic.

## License

[MIT License](LICENSE)

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request
