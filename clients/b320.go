package clients

import (
	"bytes"
	"io"

	chio "github.com/Lekuruu/chio-go"
	"github.com/Lekuruu/chio-go/internal"
)

// B320 adds partial support for multiple channels.
type B320 struct {
	*B312
}

func (client *B320) WriteMessage(stream io.Writer, message chio.Message) error {
	writer := bytes.NewBuffer([]byte{})
	internal.WriteString(writer, message.Sender)
	internal.WriteString(writer, message.Content)
	internal.WriteString(writer, message.Target)
	return client.WritePacket(stream, chio.BanchoSendMessage, writer.Bytes())
}

func (client *B320) ReadMessage(reader io.Reader) (*chio.Message, error) {
	errors := internal.NewErrorCollection()

	sender, err := internal.ReadString(reader)
	errors.Add(err)
	content, err := internal.ReadString(reader)
	errors.Add(err)
	target, err := internal.ReadString(reader)
	errors.Add(err)

	if errors.HasErrors() {
		return nil, errors.Next()
	}

	return &chio.Message{
		Sender:  sender,
		Content: content,
		Target:  target,
	}, nil
}

func (client *B320) ReadPrivateMessage(reader io.Reader) (*chio.Message, error) {
	return client.ReadMessage(reader)
}

func NewB320() *B320 {
	base := NewB312()

	client := &B320{B312: base}
	base.Instance = client
	return client
}

func init() {
	chio.RegisterClient(320, NewB320())
}
