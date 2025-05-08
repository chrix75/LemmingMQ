# LemmingMQ

LemmingMQ is a lightweight, in-memory message queue library for Go applications. It provides a simple and efficient way to implement pub/sub patterns and message dispatching within your Go applications.

## Features

- **Two Diffusion Types**:
  - **Broadcast**: Sends messages to all registered consumers
  - **Dispatch**: Distributes messages to consumers in a round-robin fashion
- **Retry Mechanism**: Configurable retry count for failed message deliveries
- **Context Support**: Respects context cancellation for graceful shutdowns
- **Minimal Dependencies**: Only depends on the Go standard library for core functionality
- **Type-Safe API**: Fully generic interfaces for reliable message handling with any data type
- **Error Handling**: Dedicated error channel for handling consumer errors

## Installation

```bash
go get github.com/chrix75/LemmingMQ
```

## Usage

### Creating a Topic

```go
package main

import (
    "fmt"
    "github.com/chrix75/LemmingMQ/topic"
)

func exampleCreateTopics() {
    // Create an error channel to receive consumer errors
    errChan := make(chan topic.TopicError[string])

    // Create a broadcast topic for string messages
    broadcastTopic := topic.NewBroadcastTopic[string](errChan)

    // Handle errors from consumers
    go func() {
        for err := range errChan {
            fmt.Printf("Error processing message %d: %v\n", err.Message.ID, err.Err)
        }
    }()
}
```

### Adding Consumers

```go
package main

import (
    "fmt"
    "github.com/chrix75/LemmingMQ/topic"
)

// Define a consumer that implements the Consumer interface
type LogConsumer struct{}

func (c LogConsumer) HandleMessage(msg topic.Message[string]) error {
    fmt.Printf("Received message: %s\n", msg.Content)
    // Process the message
    return nil
}

func exampleAddConsumer() {
    // Create an error channel
    errChan := make(chan topic.TopicError[string])

    // Create a topic for string messages
    myTopic := topic.NewBroadcastTopic[string](errChan)

    // Add a consumer to process messages
    consumer := LogConsumer{}
    myTopic.AddConsumer(consumer)
}
```

### Sending Messages

```go
package main

import (
    "fmt"
    "github.com/chrix75/LemmingMQ/topic"
)

func exampleSendMessages() {
    // Create an error channel
    errChan := make(chan topic.TopicError[string])

    // Create a topic for string messages
    myTopic := topic.NewBroadcastTopic[string](errChan)

    // Add a consumer
    myTopic.AddConsumer(&LogConsumer{})

    // Send a text message
    myTopic.SendMessage("Hello, world!")

    // Send another message
    myTopic.SendMessage("Another message")

    // Example with JSON data using a different generic type
    jsonErrChan := make(chan topic.TopicError[[]byte])
    jsonTopic := topic.NewBroadcastTopic[[]byte](jsonErrChan)

    // Send a JSON message
    jsonTopic.SendMessage([]byte(`{"greeting":"Hello, world!"}`))
}
```

### Handling Errors

LemmingMQ provides an error channel to handle errors from consumers:

```go
package main

import (
    "errors"
    "fmt"
    "math/rand"
    "github.com/chrix75/LemmingMQ/topic"
)

// Consumer that sometimes fails
type FlakyConsumer struct{}

func (c FlakyConsumer) HandleMessage(msg topic.Message[string]) error {
    // Simulate occasional failures
    if rand.Intn(10) < 3 {
        return errors.New("processing failed")
    }

    fmt.Printf("Processed message: %s\n", msg.Content)
    return nil
}

func exampleErrorHandling() {
    // Create an error channel
    errChan := make(chan topic.TopicError[string])

    // Create a topic
    myTopic := topic.NewBroadcastTopic[string](errChan)

    // Add a consumer that sometimes fails
    myTopic.AddConsumer(FlakyConsumer{})

    // Handle errors from the error channel
    go func() {
        for err := range errChan {
            fmt.Printf("Error processing message %d: %v\n", err.Message.ID, err.Err)
            // You can implement retry logic here if needed
        }
    }()

    // Send a message
    myTopic.SendMessage("Hello, world!")
}
```

## Worker Pool

The topic package uses a worker pool to manage consumers and distribute messages:

```go
package main

import (
    "fmt"
    "github.com/chrix75/LemmingMQ/topic"
)

func exampleWorkerPool() {
    // Create an error channel
    errChan := make(chan topic.TopicError[string])

    // Create a worker pool
    pool := topic.NewWorkerPool[string](errChan)

    // Create a consumer
    consumer := LogConsumer{}

    // Create a channel for the consumer
    consumerCh := make(chan topic.Message[string])

    // Start the consumer in the worker pool
    pool.startConsumer(consumerCh, consumer)

    // Send a message to all consumers
    message := topic.Message[string]{
        ID:      1,
        Content: "Hello, world!",
    }
    pool.sendMessage(message)
}
```


## API Reference

### Types

#### `topic.Message[T]`

Generic structure representing a message:
- `ID`: Unique identifier for the message
- `Content`: Content of the message with type T

#### `topic.TopicError[T]`

Generic structure representing an error from a consumer:
- `Message`: The message that caused the error
- `Err`: The error that occurred

#### `topic.Consumer[T]`

Interface for message consumers:
- `HandleMessage(Message[T]) error`: Method to handle a message

### Functions

#### `topic.NewBroadcastTopic[T](errChan chan TopicError[T]) *BroadcastTopic[T]`

Creates a new broadcast topic with the specified error channel.

#### `topic.NewWorkerPool[T](errCh chan TopicError[T]) *WorkerPool[T]`

Creates a new worker pool with the specified error channel.

### Methods

#### `(t *BroadcastTopic[T]) AddConsumer(c Consumer[T])`

Registers a new consumer to the topic.

#### `(t *BroadcastTopic[T]) SendMessage(content T)`

Sends a message with the specified content to the topic's consumers.

#### `(t *BroadcastTopic[T]) nextMessageID() int`

Generates the next message ID for the topic.

## License

[MIT License](LICENSE)

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request
