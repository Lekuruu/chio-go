package clients

import (
	"bytes"
	"io"

	chio "github.com/Lekuruu/chio-go"
	"github.com/Lekuruu/chio-go/internal"
)

// B323 changes the structure of user stats and adds the "MatchChangeBeatmap" packet.
type B323 struct {
	*B320
}

func (client *B323) ConvertInputPacketId(packetId uint16) uint16 {
	if packetId == 11 {
		// "IrcJoin" packet
		return chio.BanchoHandleIrcJoin
	}

	if packetId == 50 {
		// "MatchChangeBeatmap" packet
		return chio.OsuMatchChangeBeatmap
	}

	if packetId > 11 && packetId <= 45 {
		packetId -= 1
	}

	return packetId
}

func (client *B323) ConvertOutputPacketId(packetId uint16) uint16 {
	if packetId == chio.BanchoHandleIrcJoin {
		// "IrcJoin" packet
		return 11
	}

	if packetId == chio.OsuMatchChangeBeatmap {
		// "MatchChangeBeatmap" packet
		return 50
	}

	if packetId >= 11 && packetId < 45 {
		return packetId + 1
	}

	return packetId
}

func (client *B323) WriteUserStats(stream io.Writer, info chio.UserInfo) error {
	writer := bytes.NewBuffer([]byte{})

	if info.Presence.IsIrc {
		internal.WriteString(writer, info.Name)
		return client.WritePacket(stream, chio.BanchoHandleIrcJoin, writer.Bytes())
	}

	writeStats := info.Status.UpdateStats

	internal.WriteUint32(writer, uint32(info.Id))
	internal.WriteBoolean(writer, writeStats)

	if writeStats {
		internal.WriteString(writer, info.Name)
		internal.WriteUint64(writer, info.Stats.Rscore)
		internal.WriteFloat32(writer, float32(info.Stats.Accuracy))
		internal.WriteUint32(writer, uint32(info.Stats.Playcount))
		internal.WriteUint64(writer, info.Stats.Tscore)
		internal.WriteInt32(writer, info.Stats.Rank)
		internal.WriteString(writer, info.AvatarFilename())
		internal.WriteUint8(writer, uint8(info.Presence.Timezone+24))
		internal.WriteString(writer, info.Presence.Location())
	}

	client.WriteStatus(writer, info.Status)
	return client.WritePacket(stream, chio.BanchoHandleOsuUpdate, writer.Bytes())
}

func (client *B323) WriteUserPresence(stream io.Writer, info chio.UserInfo) error {
	if info.Presence.IsIrc {
		writer := bytes.NewBuffer([]byte{})
		internal.WriteString(writer, info.Name)
		return client.WritePacket(stream, chio.BanchoHandleIrcJoin, writer.Bytes())
	}

	// We assume that the client has not seen this user before, so
	// we send two packets: one for the user stats, and one for the "presence".
	info.Status.UpdateStats = true
	err := client.WriteUserStats(stream, info)
	if err != nil {
		return err
	}

	info.Status.UpdateStats = false
	return client.WriteUserStats(stream, info)
}

func (client *B323) ReadMatchChangeBeatmap(reader io.Reader) (*chio.Match, error) {
	return client.ReadMatch(reader)
}

func NewB323() *B323 {
	base := NewB320()
	base.SupportedPacketIds = append(base.SupportedPacketIds,
		chio.OsuMatchChangeBeatmap,
	)

	client := &B323{B320: base}
	base.Instance = client
	client.Readers[chio.OsuMatchChangeBeatmap] = internal.ReaderReadMatch()
	return client
}

func init() {
	chio.RegisterClient(323, NewB323())
}
