package topic

import (
	"context"
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCreateBroadcastTopic(t *testing.T) {
	// given
	topicName := "test"
	topicDiffusion := BroadcastTopic

	// when
	cfg := Configuration{
		Name:      topicName,
		Diffusion: topicDiffusion,
		Retries:   0,
	}

	tp := NewTopic(cfg)

	// then
	expectedTopic := Topic{
		Configuration: cfg,
	}

	assert.Equal(t, expectedTopic, *tp)
	assert.Equal(t, 0, tp.ConsumerCount())
}

func TestCreateDispatcherTopic(t *testing.T) {
	// given
	topicName := "test"
	topicDiffusion := DispatchTopic

	// when
	cfg := Configuration{
		Name:      topicName,
		Diffusion: topicDiffusion,
		Retries:   0,
	}

	tp := NewTopic(cfg)

	// then
	expectedTopic := Topic{
		Configuration: cfg,
	}

	assert.Equal(t, expectedTopic, *tp)
	assert.Equal(t, 0, tp.ConsumerCount())
}

func TestAddOneConsumerToBroadcastTopic(t *testing.T) {
	// given
	tp := NewTopic(Configuration{
		Name:      "test",
		Diffusion: BroadcastTopic,
		Retries:   0,
	})

	// when
	tp.AddConsumer(func(c context.Context, msg Message) error {
		return nil
	})

	// then
	assert.Equal(t, 1, tp.ConsumerCount())
}

func TestTopicBroadCastOneMessage(t *testing.T) {
	// given
	ctx := context.Background()

	tp := NewTopic(Configuration{
		Name:      "test",
		Diffusion: BroadcastTopic,
		Retries:   0,
	})

	var receivedContent string

	tp.AddConsumer(func(ctx context.Context, msg Message) error {
		if msg.Topic == "test" {
			receivedContent = string(msg.Content)
			return nil
		}

		return errors.New("could not encode message's content")
	})

	// when
	content := "Hello World"
	err := tp.SendMessage(ctx, []byte(content))

	// then
	assert.Nil(t, err)
	assert.Equal(t, content, receivedContent)
}

func TestTopicBroadCastOneMessageToManySubscribers(t *testing.T) {
	// given
	ctx := context.Background()

	tp := NewTopic(Configuration{
		Name:      "test",
		Diffusion: BroadcastTopic,
		Retries:   0,
	})

	receivedContent := make([]string, 0)

	consumerCallback := func(consumerID int) ConsumerCallback {
		return func(ctx context.Context, msg Message) error {
			if msg.Topic == "test" {
				content := fmt.Sprintf("%d>%s", consumerID, msg.Content)
				receivedContent = append(receivedContent, content)
				return nil
			}

			return errors.New("could not encode message's content")
		}
	}

	tp.AddConsumer(consumerCallback(1))
	tp.AddConsumer(consumerCallback(2))

	// when
	content := "Hello World"
	err := tp.SendMessage(ctx, []byte(content))

	// then
	assert.Nil(t, err)
	expectedContent := []string{"1>Hello World", "2>Hello World"}
	expectedIDs := []int{1, 1}
	assert.Equal(t, expectedContent, receivedContent)
	assert.Equal(t, expectedIDs, expectedIDs)
}

func TestTopicDispatchOneMessageToManySubscribers(t *testing.T) {
	// given
	ctx := context.Background()

	tp := NewTopic(Configuration{
		Name:      "test",
		Diffusion: DispatchTopic,
		Retries:   0,
	})

	receivedContent := make([]string, 0)
	messageIDs := make([]int, 0)

	consumerCallback := func(ctx context.Context, msg Message) error {
		if msg.Topic == "test" {
			receivedContent = append(receivedContent, string(msg.Content))
			messageIDs = append(messageIDs, msg.ID)
			return nil
		}

		return errors.New("could not encode message's content")
	}

	tp.AddConsumer(consumerCallback)
	tp.AddConsumer(consumerCallback)

	// when
	content := "Hello World"
	err := tp.SendMessage(ctx, []byte(content))

	// then
	assert.Nil(t, err)
	expectedContent := []string{"Hello World"}
	expectedIDs := []int{1}
	assert.Equal(t, expectedContent, receivedContent)
	assert.Equal(t, expectedIDs, messageIDs)
}

func TestTopicDispatchTwoMessagesToManySubscribers(t *testing.T) {
	// given
	ctx := context.Background()

	tp := NewTopic(Configuration{
		Name:      "test",
		Diffusion: DispatchTopic,
		Retries:   0,
	})

	receivedContent := make([]string, 0)
	messageIDs := make([]int, 0)

	consumerCallback := func(consumerID int) ConsumerCallback {
		return func(ctx context.Context, msg Message) error {
			if msg.Topic == "test" {
				content := fmt.Sprintf("%d>%s", consumerID, msg.Content)
				receivedContent = append(receivedContent, content)
				messageIDs = append(messageIDs, msg.ID)
				return nil
			}

			return errors.New("could not encode message's content")
		}
	}

	tp.AddConsumer(consumerCallback(1))
	tp.AddConsumer(consumerCallback(2))

	// when
	content := "Hello World"
	err := tp.SendMessage(ctx, []byte(content))
	if err != nil {
		t.Fatal(err)
	}

	err = tp.SendMessage(ctx, []byte(content))

	// then
	assert.Nil(t, err)
	expectedContent := []string{"1>Hello World", "2>Hello World"}
	expectedIDs := []int{1, 2}

	assert.Equal(t, expectedContent, receivedContent)
	assert.Equal(t, expectedIDs, messageIDs)
}

func TestBroadcastWithRetry(t *testing.T) {
	// given
	ctx := context.Background()

	tp := NewTopic(Configuration{
		Name:      "test",
		Diffusion: BroadcastTopic,
		Retries:   1,
	})

	receivedContent := make([]string, 0)
	messageIDs := make([]int, 0)

	consumerCallback := func(ctx context.Context, msg Message) error {
		if msg.Topic == "test" {
			receivedContent = append(receivedContent, string(msg.Content))
			messageIDs = append(messageIDs, msg.ID)
			return errors.New("simulated error")
		}

		return errors.New("could not encode message's content")
	}

	tp.AddConsumer(consumerCallback)

	// when
	content := "Hello World"
	err := tp.SendMessage(ctx, []byte(content))

	// then
	assert.NotNil(t, err)
	expectedContent := []string{"Hello World", "Hello World"}
	expectedIDs := []int{1, 1}

	assert.Equal(t, expectedContent, receivedContent)
	assert.Equal(t, expectedIDs, messageIDs)
}

func TestDispatchWithRetry(t *testing.T) {
	// given
	ctx := context.Background()

	tp := NewTopic(Configuration{
		Name:      "test",
		Diffusion: DispatchTopic,
		Retries:   1,
	})

	receivedContent := make([]string, 0)
	messageIDs := make([]int, 0)

	consumerCallback := func(ctx context.Context, msg Message) error {
		if msg.Topic == "test" {
			receivedContent = append(receivedContent, string(msg.Content))
			messageIDs = append(messageIDs, msg.ID)
			return errors.New("simulated error")
		}

		return errors.New("could not encode message's content")
	}

	tp.AddConsumer(consumerCallback)

	// when
	content := "Hello World"
	err := tp.SendMessage(ctx, []byte(content))

	// then
	assert.NotNil(t, err)
	expectedContent := []string{"Hello World", "Hello World"}
	expectedIDs := []int{1, 1}

	assert.Equal(t, expectedContent, receivedContent)
	assert.Equal(t, expectedIDs, messageIDs)
}

func TestCancelMessageSending(t *testing.T) {
	// given
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tp := NewTopic(Configuration{
		Name:      "test",
		Diffusion: DispatchTopic,
		Retries:   1,
	})

	consumerCallback := func(ctx context.Context, msg Message) error {
		return nil
	}

	tp.AddConsumer(consumerCallback)

	// when
	cancel()

	content := "Hello World"
	err := tp.SendMessage(ctx, []byte(content))

	// then
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "context canceled")
}
