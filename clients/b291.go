package clients

import (
	"bytes"
	"io"

	chio "github.com/Lekuruu/chio-go"
	"github.com/Lekuruu/chio-go/internal"
)

// B291 implements the GetAttention & Announce packets.
type B291 struct {
	*B282
}

func (client *B291) WriteGetAttention(stream io.Writer) error {
	return client.WritePacket(stream, chio.BanchoGetAttention, []byte{})
}

func (client *B291) WriteAnnouncement(stream io.Writer, message string) error {
	writer := bytes.NewBuffer([]byte{})
	internal.WriteString(writer, message)
	return client.WritePacket(stream, chio.BanchoAnnounce, writer.Bytes())
}

func (client *B291) WriteRestart(stream io.Writer, retryMs int32) error {
	// NOTE: This is a backport of the actual restart packet, that
	// simply announces the server restart to the user.
	return client.WriteAnnouncement(stream, "Bancho is restarting, please wait...")
}

func NewB291() *B291 {
	base := NewB282()
	base.SupportedPacketIds = append(base.SupportedPacketIds,
		chio.BanchoGetAttention,
		chio.BanchoAnnounce,
	)

	return &B291{B282: base}
}

func init() {
	chio.RegisterClient(291, NewB291())
}
