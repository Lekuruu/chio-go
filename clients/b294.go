package clients

import (
	"bytes"
	"fmt"
	"io"

	chio "github.com/Lekuruu/chio-go"
	"github.com/Lekuruu/chio-go/internal"
)

// B294 implements private messages, as well as score frames in spectating.
type B294 struct {
	*B291
}

func (client *B294) WriteMessage(stream io.Writer, message chio.Message) error {
	writer := bytes.NewBuffer([]byte{})
	internal.WriteString(writer, message.Sender)
	internal.WriteString(writer, message.Content)

	isDirectMessage := message.Target != "#osu"
	internal.WriteBoolean(writer, isDirectMessage)

	return client.WritePacket(stream, chio.BanchoSendMessage, writer.Bytes())
}

func (client *B294) ReadPrivateMessage(reader io.Reader) (*chio.Message, error) {
	target, err := internal.ReadString(reader)
	if err != nil {
		return nil, err
	}
	content, err := internal.ReadString(reader)
	if err != nil {
		return nil, err
	}
	isDirectMessage, err := internal.ReadBoolean(reader)
	if err != nil {
		return nil, err
	}

	if !isDirectMessage {
		return nil, fmt.Errorf("expected direct message, got channel message")
	}

	return &chio.Message{Sender: "", Content: content, Target: target, SenderId: 0}, nil
}

func (client *B294) WriteSpectateFrames(stream io.Writer, bundle chio.ReplayFrameBundle) error {
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

func (client *B294) ReadFrameBundle(reader io.Reader) (*chio.ReplayFrameBundle, error) {
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

func (client *B294) WriteScoreFrame(writer io.Writer, frame *chio.ScoreFrame) {
	internal.WriteString(writer, frame.Checksum())
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

func (client *B294) ReadScoreFrame(reader io.Reader) (*chio.ScoreFrame, error) {
	_, err := internal.ReadString(reader)
	if err != nil {
		return nil, err
	}
	id, err := internal.ReadUint8(reader)
	if err != nil {
		return nil, err
	}
	p300, err := internal.ReadUint16(reader)
	if err != nil {
		return nil, err
	}
	p100, err := internal.ReadUint16(reader)
	if err != nil {
		return nil, err
	}
	p50, err := internal.ReadUint16(reader)
	if err != nil {
		return nil, err
	}
	geki, err := internal.ReadUint16(reader)
	if err != nil {
		return nil, err
	}
	katu, err := internal.ReadUint16(reader)
	if err != nil {
		return nil, err
	}
	miss, err := internal.ReadUint16(reader)
	if err != nil {
		return nil, err
	}
	score, err := internal.ReadUint32(reader)
	if err != nil {
		return nil, err
	}
	maxCombo, err := internal.ReadUint16(reader)
	if err != nil {
		return nil, err
	}
	currentCombo, err := internal.ReadUint16(reader)
	if err != nil {
		return nil, err
	}
	perfect, err := internal.ReadBoolean(reader)
	if err != nil {
		return nil, err
	}
	hp, err := internal.ReadUint8(reader)
	if err != nil {
		return nil, err
	}

	return &chio.ScoreFrame{
		Time:         0,
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

func NewB294() *B294 {
	base := NewB291()
	base.SupportedPacketIds = append(base.SupportedPacketIds, chio.OsuSendIrcMessagePrivate)

	client := &B294{B291: base}
	base.Instance = client
	client.Readers[chio.OsuSendIrcMessagePrivate] = internal.ReaderReadPrivateMessage()
	client.Readers[chio.OsuSpectateFrames] = internal.ReaderReadFrameBundle()
	return client
}

func init() {
	chio.RegisterClient(294, NewB294())
}
