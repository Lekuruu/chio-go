package clients

import (
	"bytes"
	"io"

	chio "github.com/Lekuruu/chio-go"
	"github.com/Lekuruu/chio-go/internal"
)

// B296 adds the "Time" value to score frames.
type B296 struct {
	*B294
}

func (client *B296) WriteScoreFrame(writer io.Writer, frame *chio.ScoreFrame) {
	internal.WriteString(writer, frame.Checksum())
	internal.WriteInt32(writer, frame.Time)
	internal.WriteUint8(writer, frame.Id)
	internal.WriteUint16(writer, frame.Total300)
	internal.WriteUint16(writer, frame.Total100)
	internal.WriteUint16(writer, frame.Total50)
	internal.WriteUint16(writer, frame.TotalGeki)
	internal.WriteUint16(writer, frame.TotalKatu)
	internal.WriteUint16(writer, frame.TotalMiss)
	internal.WriteUint32(writer, frame.TotalScore)
	internal.WriteUint16(writer, frame.MaxCombo)
	internal.WriteUint16(writer, frame.CurrentCombo)
	internal.WriteBoolean(writer, frame.Perfect)
	internal.WriteUint8(writer, frame.Hp)
}

func (client *B296) ReadScoreFrame(reader io.Reader) (*chio.ScoreFrame, error) {
	errors := internal.NewErrorCollection()

	_, err := internal.ReadString(reader)
	errors.Add(err)
	time, err := internal.ReadInt32(reader)
	errors.Add(err)
	id, err := internal.ReadUint8(reader)
	errors.Add(err)
	p300, err := internal.ReadUint16(reader)
	errors.Add(err)
	p100, err := internal.ReadUint16(reader)
	errors.Add(err)
	p50, err := internal.ReadUint16(reader)
	errors.Add(err)
	geki, err := internal.ReadUint16(reader)
	errors.Add(err)
	katu, err := internal.ReadUint16(reader)
	errors.Add(err)
	miss, err := internal.ReadUint16(reader)
	errors.Add(err)
	score, err := internal.ReadUint32(reader)
	errors.Add(err)
	maxCombo, err := internal.ReadUint16(reader)
	errors.Add(err)
	currentCombo, err := internal.ReadUint16(reader)
	errors.Add(err)
	perfect, err := internal.ReadBoolean(reader)
	errors.Add(err)
	hp, err := internal.ReadUint8(reader)
	errors.Add(err)

	if errors.HasErrors() {
		return nil, errors.Next()
	}

	return &chio.ScoreFrame{
		Time:         time,
		Id:           id,
		Total300:     p300,
		Total100:     p100,
		Total50:      p50,
		TotalGeki:    geki,
		TotalKatu:    katu,
		TotalMiss:    miss,
		TotalScore:   score,
		MaxCombo:     maxCombo,
		CurrentCombo: currentCombo,
		Perfect:      perfect,
		Hp:           hp,
		TagByte:      0,
	}, nil
}

func (client *B296) WriteSpectateFrames(stream io.Writer, bundle chio.ReplayFrameBundle) error {
	writer := bytes.NewBuffer([]byte{})
	internal.WriteUint16(writer, uint16(len(bundle.Frames)))

	for _, frame := range bundle.Frames {
		leftMouse := chio.ButtonStateLeft1&frame.ButtonState > 0 || chio.ButtonStateLeft2&frame.ButtonState > 0
		rightMouse := chio.ButtonStateRight1&frame.ButtonState > 0 || chio.ButtonStateRight2&frame.ButtonState > 0

		internal.WriteBoolean(writer, leftMouse)
		internal.WriteBoolean(writer, rightMouse)
		internal.WriteFloat32(writer, frame.MouseX)
		internal.WriteFloat32(writer, frame.MouseY)
		internal.WriteInt32(writer, frame.Time)
	}

	internal.WriteUint8(writer, bundle.Action)

	if bundle.Frame != nil {
		client.WriteScoreFrame(writer, bundle.Frame)
	}

	return client.WritePacket(stream, chio.BanchoSpectateFrames, writer.Bytes())
}

func (client *B296) ReadFrameBundle(reader io.Reader) (*chio.ReplayFrameBundle, error) {
	count, err := internal.ReadUint16(reader)
	if err != nil {
		return nil, err
	}

	frames := make([]*chio.ReplayFrame, count)
	for i := 0; i < int(count); i++ {
		frame, err := client.ReadReplayFrame(reader)
		if err != nil {
			return nil, err
		}
		frames[i] = frame
	}

	action, err := internal.ReadUint8(reader)
	if err != nil {
		return nil, err
	}

	scoreFrame, err := client.ReadScoreFrame(reader)
	if err != nil {
		scoreFrame = nil
	}

	return &chio.ReplayFrameBundle{Frames: frames, Action: action, Frame: scoreFrame}, nil
}

func NewB296() *B296 {
	base := NewB294()

	client := &B296{B294: base}
	client.Readers[chio.OsuSpectateFrames] = internal.ReaderReadFrameBundle()
	return client
}

func init() {
	chio.RegisterClient(296, NewB296())
}
