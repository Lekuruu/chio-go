package clients

import (
	"bytes"
	"io"

	chio "github.com/Lekuruu/chio-go"
	"github.com/Lekuruu/chio-go/internal"
)

// B298 adds a partial implementation of multiplayer, as well as fellow spectators.
type B298 struct {
	*B296
}

func (client *B298) WriteMatchUpdate(stream io.Writer, match chio.Match) error {
	if match.Id > 0xFF {
		// Match IDs greater than 255 are not supported in this client
		return nil
	}
	return client.WritePacket(stream, chio.BanchoMatchUpdate, client.WriteMatch(match))
}

func (client *B298) WriteMatchNew(stream io.Writer, match chio.Match) error {
	if match.Id > 0xFF {
		// Match IDs greater than 255 are not supported in this client
		return nil
	}
	return client.WritePacket(stream, chio.BanchoMatchNew, client.WriteMatch(match))
}

func (client *B298) WriteMatchDisband(stream io.Writer, matchId int32) error {
	writer := bytes.NewBuffer([]byte{})
	internal.WriteInt32(writer, matchId)
	return client.WritePacket(stream, chio.BanchoMatchDisband, writer.Bytes())
}

func (client *B298) WriteLobbyJoin(stream io.Writer, userId int32) error {
	writer := bytes.NewBuffer([]byte{})
	internal.WriteInt32(writer, userId)
	return client.WritePacket(stream, chio.BanchoLobbyJoin, writer.Bytes())
}

func (client *B298) WriteLobbyPart(stream io.Writer, userId int32) error {
	writer := bytes.NewBuffer([]byte{})
	internal.WriteInt32(writer, userId)
	return client.WritePacket(stream, chio.BanchoLobbyPart, writer.Bytes())
}

func (client *B298) WriteMatchJoinSuccess(stream io.Writer, match chio.Match) error {
	return client.WritePacket(stream, chio.BanchoMatchJoinSuccess, client.WriteMatch(match))
}

func (client *B298) WriteMatchJoinFail(stream io.Writer) error {
	return client.WritePacket(stream, chio.BanchoMatchJoinFail, []byte{})
}

func (client *B298) WriteFellowSpectatorJoined(stream io.Writer, userId int32) error {
	writer := bytes.NewBuffer([]byte{})
	internal.WriteInt32(writer, userId)
	return client.WritePacket(stream, chio.BanchoFellowSpectatorJoined, writer.Bytes())
}

func (client *B298) WriteFellowSpectatorLeft(stream io.Writer, userId int32) error {
	writer := bytes.NewBuffer([]byte{})
	internal.WriteInt32(writer, userId)
	return client.WritePacket(stream, chio.BanchoFellowSpectatorLeft, writer.Bytes())
}

func (client *B298) WriteMatch(match chio.Match) []byte {
	slotSize := client.MatchSlotSize()

	slotsOpen := make([]bool, slotSize)
	slotsUsed := make([]bool, slotSize)
	slotsReady := make([]bool, slotSize)

	for i := 0; i < slotSize; i++ {
		slotsOpen[i] = match.Slots[i].Status == chio.SlotStatusOpen
		slotsUsed[i] = match.Slots[i].HasPlayer()
		slotsReady[i] = match.Slots[i].Status == chio.SlotStatusReady
	}

	writer := bytes.NewBuffer([]byte{})
	internal.WriteUint8(writer, uint8(match.Id))
	internal.WriteUint8(writer, match.Type)
	internal.WriteString(writer, match.Name)
	internal.WriteString(writer, match.BeatmapText)
	internal.WriteInt32(writer, match.BeatmapId)
	internal.WriteString(writer, match.BeatmapChecksum)
	internal.WriteBoolList(writer, slotsOpen)
	internal.WriteBoolList(writer, slotsUsed)
	internal.WriteBoolList(writer, slotsReady)

	for i := 0; i < slotSize; i++ {
		if match.Slots[i].HasPlayer() {
			internal.WriteInt32(writer, match.Slots[i].UserId)
		}
	}

	return writer.Bytes()
}

func (client *B298) ReadMatch(reader io.Reader) (*chio.Match, error) {
	slotSize := client.MatchSlotSize()
	errors := internal.NewErrorCollection()

	matchId, err := internal.ReadUint8(reader)
	errors.Add(err)
	matchType, err := internal.ReadUint8(reader)
	errors.Add(err)
	name, err := internal.ReadString(reader)
	errors.Add(err)
	beatmapText, err := internal.ReadString(reader)
	errors.Add(err)
	beatmapId, err := internal.ReadInt32(reader)
	errors.Add(err)
	beatmapChecksum, err := internal.ReadString(reader)
	errors.Add(err)

	slotsOpen, err := internal.ReadBoolList(reader)
	errors.Add(err)
	slotsUsed, err := internal.ReadBoolList(reader)
	errors.Add(err)
	slotsReady, err := internal.ReadBoolList(reader)
	errors.Add(err)

	if errors.HasErrors() {
		return nil, errors.Next()
	}

	slots := make([]*chio.MatchSlot, slotSize)

	for i := 0; i < slotSize; i++ {
		slot := &chio.MatchSlot{}

		if slotsOpen[i] {
			slot.Status = chio.SlotStatusOpen
		} else {
			slot.Status = chio.SlotStatusLocked
		}

		if slotsUsed[i] {
			slot.Status = chio.SlotStatusNotReady
		}

		if slotsReady[i] {
			slot.Status = chio.SlotStatusReady
		}

		if slot.HasPlayer() {
			userId, err := internal.ReadInt32(reader)
			if err != nil {
				return nil, err
			}
			slot.UserId = userId
		}

		slots[i] = slot
	}

	return &chio.Match{
		Id:              int32(matchId),
		Type:            matchType,
		Name:            name,
		BeatmapText:     beatmapText,
		BeatmapId:       beatmapId,
		BeatmapChecksum: beatmapChecksum,
		Slots:           slots,
	}, nil
}

func (client *B298) ReadMatchJoin(reader io.Reader) (*chio.MatchJoin, error) {
	matchId, err := internal.ReadInt32(reader)
	if err != nil {
		return nil, err
	}
	return &chio.MatchJoin{MatchId: matchId}, nil
}

func (client *B298) ReadMatchChangeSlot(reader io.Reader) (int32, error) {
	return internal.ReadInt32(reader)
}

func (client *B298) ReadMatchLock(reader io.Reader) (int32, error) {
	return internal.ReadInt32(reader)
}

func NewB298() *B298 {
	base := NewB296()
	base.SupportedPacketIds = append(base.SupportedPacketIds,
		chio.BanchoMatchUpdate,
		chio.BanchoMatchNew,
		chio.BanchoMatchDisband,
		chio.OsuLobbyPart,
		chio.OsuLobbyJoin,
		chio.OsuMatchCreate,
		chio.OsuMatchJoin,
		chio.OsuMatchPart,
		chio.BanchoLobbyJoin,
		chio.BanchoLobbyPart,
		chio.BanchoMatchJoinSuccess,
		chio.BanchoMatchJoinFail,
		chio.OsuMatchChangeSlot,
		chio.OsuMatchReady,
		chio.OsuMatchLock,
		chio.OsuMatchChangeSettings,
		chio.BanchoFellowSpectatorJoined,
		chio.BanchoFellowSpectatorLeft,
	)

	client := &B298{B296: base}
	base.Instance = client

	// Register packet readers
	client.Readers[chio.OsuLobbyJoin] = internal.ReaderReadEmpty()
	client.Readers[chio.OsuLobbyPart] = internal.ReaderReadEmpty()
	client.Readers[chio.OsuMatchCreate] = internal.ReaderReadMatch()
	client.Readers[chio.OsuMatchJoin] = internal.ReaderReadMatchJoin()
	client.Readers[chio.OsuMatchPart] = internal.ReaderReadEmpty()
	client.Readers[chio.OsuMatchChangeSlot] = internal.ReaderReadMatchChangeSlot()
	client.Readers[chio.OsuMatchReady] = internal.ReaderReadEmpty()
	client.Readers[chio.OsuMatchLock] = internal.ReaderReadMatchLock()
	client.Readers[chio.OsuMatchChangeSettings] = internal.ReaderReadMatch()

	return client
}

func init() {
	chio.RegisterClient(298, NewB298())
}
