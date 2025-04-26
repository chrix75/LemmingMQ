package topic

type Message struct {
	ContentType string
	Content     []byte
	ID          int
}

// NewMessage creates a new Message with the specified ID, content type, and content.
// This function is typically used internally by the Topic when sending messages.
//
// Example:
//
//	// Create a text message with ID 1
//	textMsg := NewMessage(1, "text/plain", []byte("Hello, world!"))
//
//	// Create a JSON message with ID 2
//	jsonMsg := NewMessage(2, "application/json", []byte(`{"greeting":"Hello, world!"}`))
func NewMessage(id int, contentType string, content []byte) Message {
	return Message{
		ID:          id,
		ContentType: contentType,
		Content:     content,
	}
}
