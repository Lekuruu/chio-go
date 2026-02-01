package clients

import (
	"bytes"
	"io"

	chio "github.com/Lekuruu/chio-go"
	"github.com/Lekuruu/chio-go/internal"
)

// B312 adds the match start & update packets, as well
// as the "InProgress" boolean inside the match struct.
type B312 struct {
	*B298
}

func (client *B312) WriteMatchStart(stream io.Writer, match chio.Match) error {
	return client.WritePacket(stream, chio.BanchoMatchStart, []byte{})
}

func (client *B312) WriteMatchScoreUpdate(stream io.Writer, frame chio.ScoreFrame) error {
	writer := bytes.NewBuffer([]byte{})
	client.WriteScoreFrame(writer, &frame)
	return client.WritePacket(stream, chio.BanchoMatchScoreUpdate, writer.Bytes())
}

func (client *B312) WriteMatch(match chio.Match) []byte {
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
	internal.WriteBoolean(writer, match.InProgress)
	internal.WriteUint8(writer, match.Type)
	internal.WriteString(writer, match.Name)
	internal.WriteString(writer, match.BeatmapText)
	internal.WriteInt32(writer, match.BeatmapId)
	internal.WriteString(writer, match.BeatmapChecksum)
	internal.WriteBoolList(writer, slotsOpen, client.Instance.MatchSlotSize())
	internal.WriteBoolList(writer, slotsUsed, client.Instance.MatchSlotSize())
	internal.WriteBoolList(writer, slotsReady, client.Instance.MatchSlotSize())

	for i := 0; i < slotSize; i++ {
		if match.Slots[i].HasPlayer() {
			internal.WriteInt32(writer, match.Slots[i].UserId)
		}
	}

	return writer.Bytes()
}

func (client *B312) ReadMatch(reader io.Reader) (*chio.Match, error) {
	slotSize := client.MatchSlotSize()
	errors := internal.NewErrorCollection()

	matchId, err := internal.ReadUint8(reader)
	errors.Add(err)
	inProgress, err := internal.ReadBoolean(reader)
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

	slotsOpen, err := internal.ReadBoolList(reader, client.Instance.MatchSlotSize())
	errors.Add(err)
	slotsUsed, err := internal.ReadBoolList(reader, client.Instance.MatchSlotSize())
	errors.Add(err)
	slotsReady, err := internal.ReadBoolList(reader, client.Instance.MatchSlotSize())
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
		InProgress:      inProgress,
		Type:            matchType,
		Name:            name,
		BeatmapText:     beatmapText,
		BeatmapId:       beatmapId,
		BeatmapChecksum: beatmapChecksum,
		Slots:           slots,
	}, nil
}

func (client *B312) ReadMatchScoreUpdate(reader io.Reader) (*chio.ScoreFrame, error) {
	return client.ReadScoreFrame(reader)
}

func NewB312() *B312 {
	base := NewB298()
	base.SupportedPacketIds = append(base.SupportedPacketIds,
		chio.OsuMatchStart,
		chio.BanchoMatchStart,
		chio.OsuMatchScoreUpdate,
		chio.BanchoMatchScoreUpdate,
		chio.OsuMatchComplete,
	)

	client := &B312{B298: base}
	base.Instance = client
	client.Readers[chio.OsuMatchStart] = internal.ReaderReadEmpty()
	client.Readers[chio.OsuMatchScoreUpdate] = internal.ReaderReadScoreFrame()
	client.Readers[chio.OsuMatchComplete] = internal.ReaderReadEmpty()
	return client
}

func init() {
	chio.RegisterClient(312, NewB312())
}
