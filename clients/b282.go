package clients

import (
	"bytes"
	"fmt"
	"io"

	chio "github.com/Lekuruu/chio-go"
	"github.com/Lekuruu/chio-go/internal"
)

// B282 is the initial implementation of the bancho protocol.
// Every following version will be based on it.
type B282 struct {
	chio.BanchoIO
	SupportedPacketIds []uint16
	ProtocolVer        int
	SlotSize           int
	Readers            chio.ReaderRegistry
}

func (client *B282) WritePacket(stream io.Writer, packetId uint16, data []byte) error {
	// Convert packetId back for the client
	packetId = client.ConvertOutputPacketId(packetId)
	writer := bytes.NewBuffer([]byte{})

	err := internal.WriteUint16(writer, packetId)
	if err != nil {
		return err
	}

	compressed := internal.CompressData(data)
	err = internal.WriteUint32(writer, uint32(len(compressed)))
	if err != nil {
		return err
	}

	_, err = writer.Write(compressed)
	if err != nil {
		return err
	}

	_, err = stream.Write(writer.Bytes())
	return err
}

func (client *B282) ReadPacket(stream io.Reader) (packet *chio.BanchoPacket, err error) {
	packet = &chio.BanchoPacket{}
	packet.Id, err = internal.ReadUint16(stream)
	if err != nil {
		return nil, err
	}

	// Convert packet ID to a usable value
	packet.Id = client.ConvertInputPacketId(packet.Id)

	if !client.ImplementsPacket(packet.Id) {
		return nil, fmt.Errorf("packet '%d' not implemented", packet.Id)
	}

	length, err := internal.ReadInt32(stream)
	if err != nil {
		return nil, err
	}

	compressedData := make([]byte, length)
	n, err := stream.Read(compressedData)
	if err != nil {
		return nil, err
	}

	if n != int(length) {
		return nil, fmt.Errorf("expected %d bytes, got %d", length, n)
	}

	data, err := internal.DecompressData(compressedData)
	if err != nil {
		return nil, err
	}

	reader, ok := client.Readers[packet.Id]
	packet.Data = nil

	if ok {
		packet.Data, err = reader(client, bytes.NewReader(data))
		if err != nil {
			return nil, err
		}
	}

	return packet, nil
}

func (client *B282) SupportedPackets() []uint16 {
	return client.SupportedPacketIds
}

func (client *B282) ImplementsPacket(packetId uint16) bool {
	for _, id := range client.SupportedPackets() {
		if id == packetId {
			return true
		}
	}
	return false
}

func (client *B282) ProtocolVersion() int {
	return client.ProtocolVer
}

func (client *B282) OverrideProtocolVersion(version int) {
	client.ProtocolVer = version
}

func (client *B282) MatchSlotSize() int {
	return client.SlotSize
}

func (client *B282) OverrideMatchSlotSize(amount int) {
	client.SlotSize = amount
}

func (client *B282) ConvertInputPacketId(packetId uint16) uint16 {
	if packetId == 11 {
		// "IrcJoin" packet
		return chio.BanchoHandleIrcJoin
	}
	if packetId > 11 && packetId <= 45 {
		packetId -= 1
	}
	if packetId > 50 {
		packetId -= 1
	}
	return packetId
}

func (client *B282) ConvertOutputPacketId(packetId uint16) uint16 {
	if packetId == chio.BanchoHandleIrcJoin {
		// "IrcJoin" packet
		return 11
	}
	if packetId >= 11 && packetId < 45 {
		return packetId + 1
	}
	if packetId > 50 {
		packetId += 1
	}
	return packetId
}

func (client *B282) GetReaders() chio.ReaderRegistry {
	return client.Readers
}

func (client *B282) WriteLoginReply(stream io.Writer, reply int32) error {
	writer := bytes.NewBuffer([]byte{})
	internal.WriteInt32(writer, reply)
	return client.WritePacket(stream, chio.BanchoLoginReply, writer.Bytes())
}

func (client *B282) WriteMessage(stream io.Writer, message chio.Message) error {
	if message.Target != "#osu" {
		// Private messages & channels have not been implemented yet
		return nil
	}

	writer := bytes.NewBuffer([]byte{})
	internal.WriteString(writer, message.Sender)
	internal.WriteString(writer, message.Content)
	return client.WritePacket(stream, chio.BanchoSendMessage, writer.Bytes())
}

func (client *B282) WritePing(stream io.Writer) error {
	return client.WritePacket(stream, chio.BanchoPing, []byte{})
}

func (client *B282) WriteIrcChangeUsername(stream io.Writer, oldName string, newName string) error {
	writer := bytes.NewBuffer([]byte{})
	internal.WriteString(writer, fmt.Sprintf("%s>>>>%s", oldName, newName))
	return client.WritePacket(stream, chio.BanchoHandleIrcChangeUsername, writer.Bytes())
}

func (client *B282) WriteUserStats(stream io.Writer, info chio.UserInfo) error {
	writer := bytes.NewBuffer([]byte{})

	if info.Presence.IsIrc {
		internal.WriteString(writer, info.Name)
		return client.WritePacket(stream, chio.BanchoHandleIrcJoin, writer.Bytes())
	}

	client.WriteStats(writer, info)
	return client.WritePacket(stream, chio.BanchoHandleOsuUpdate, writer.Bytes())
}

func (client *B282) WriteUserQuit(stream io.Writer, quit chio.UserQuit) error {
	writer := bytes.NewBuffer([]byte{})

	if quit.Info.Presence.IsIrc && quit.QuitState != chio.QuitStateIrcRemaining {
		internal.WriteString(writer, quit.Info.Name)
		return client.WritePacket(stream, chio.BanchoHandleIrcQuit, writer.Bytes())
	}

	if quit.QuitState == chio.QuitStateOsuRemaining {
		return nil
	}

	client.WriteStats(writer, *quit.Info)
	return client.WritePacket(stream, chio.BanchoHandleOsuQuit, writer.Bytes())
}

func (client *B282) WriteSpectatorJoined(stream io.Writer, userId int32) error {
	writer := bytes.NewBuffer([]byte{})
	internal.WriteInt32(writer, userId)
	return client.WritePacket(stream, chio.BanchoSpectatorJoined, writer.Bytes())
}

func (client *B282) WriteSpectatorLeft(stream io.Writer, userId int32) error {
	writer := bytes.NewBuffer([]byte{})
	internal.WriteInt32(writer, userId)
	return client.WritePacket(stream, chio.BanchoSpectatorLeft, writer.Bytes())
}

func (client *B282) WriteSpectateFrames(stream io.Writer, bundle chio.ReplayFrameBundle) error {
	writer := bytes.NewBuffer([]byte{})
	internal.WriteUint16(writer, uint16(len(bundle.Frames)))

	for _, frame := range bundle.Frames {
		// Convert button state
		leftMouse := chio.ButtonStateLeft1&frame.ButtonState > 0 || chio.ButtonStateLeft2&frame.ButtonState > 0
		rightMouse := chio.ButtonStateRight1&frame.ButtonState > 0 || chio.ButtonStateRight2&frame.ButtonState > 0

		internal.WriteBoolean(writer, leftMouse)
		internal.WriteBoolean(writer, rightMouse)
		internal.WriteFloat32(writer, frame.MouseX)
		internal.WriteFloat32(writer, frame.MouseY)
		internal.WriteInt32(writer, frame.Time)
	}

	internal.WriteUint8(writer, bundle.Action)
	return client.WritePacket(stream, chio.BanchoSpectateFrames, writer.Bytes())
}

func (client *B282) WriteVersionUpdate(stream io.Writer) error {
	return client.WritePacket(stream, chio.BanchoVersionUpdate, []byte{})
}

func (client *B282) WriteSpectatorCantSpectate(stream io.Writer, userId int32) error {
	writer := bytes.NewBuffer([]byte{})
	internal.WriteInt32(writer, userId)
	return client.WritePacket(stream, chio.BanchoSpectatorCantSpectate, writer.Bytes())
}

func (client *B282) WriteStatus(writer io.Writer, status *chio.UserStatus) error {
	// Convert action enum
	action := status.Action

	if status.UpdateStats {
		// This will make the client update the user's stats
		// It will not be present in later versions
		action = chio.StatusStatsUpdate
	}

	internal.WriteUint8(writer, action)

	if action != chio.StatusUnknown {
		internal.WriteString(writer, status.Text)
		internal.WriteString(writer, status.BeatmapChecksum)
		internal.WriteUint16(writer, uint16(status.Mods))
	}

	return nil
}

func (client *B282) WriteStats(writer io.Writer, info chio.UserInfo) error {
	internal.WriteInt32(writer, info.Id)
	internal.WriteString(writer, info.Name)
	internal.WriteUint64(writer, info.Stats.Rscore)
	internal.WriteFloat64(writer, info.Stats.Accuracy)
	internal.WriteInt32(writer, info.Stats.Playcount)
	internal.WriteUint64(writer, info.Stats.Tscore)
	internal.WriteInt32(writer, info.Stats.Rank)
	internal.WriteString(writer, info.AvatarFilename())
	client.WriteStatus(writer, info.Status)
	internal.WriteUint8(writer, uint8(info.Presence.Timezone+24))
	internal.WriteString(writer, info.Presence.Location())
	return nil
}

// Redirect UserPresence packets to UserStats
func (client *B282) WriteUserPresence(stream io.Writer, info chio.UserInfo) error {
	return client.WriteUserStats(stream, info)
}

func (client *B282) WriteUserPresenceSingle(stream io.Writer, info chio.UserInfo) error {
	return client.WriteUserPresence(stream, info)
}

func (client *B282) WriteUserPresenceBundle(stream io.Writer, infos []chio.UserInfo) error {
	for _, info := range infos {
		err := client.WriteUserPresence(stream, info)
		if err != nil {
			return err
		}
	}
	return nil
}

func (client *B282) ReadStatus(reader io.Reader) (*chio.UserStatus, error) {
	var err error
	errors := internal.NewErrorCollection()
	status := &chio.UserStatus{}
	status.Action, err = internal.ReadUint8(reader)
	errors.Add(err)

	if status.Action != chio.StatusUnknown {
		status.Text, err = internal.ReadString(reader)
		errors.Add(err)
		status.BeatmapChecksum, err = internal.ReadString(reader)
		errors.Add(err)
		mods, err := internal.ReadUint16(reader)
		errors.Add(err)
		status.Mods = uint32(mods)
	}

	return status, errors.Next()
}

func (client *B282) ReadMessage(reader io.Reader) (*chio.Message, error) {
	var err error
	message := &chio.Message{}
	message.Content, err = internal.ReadString(reader)
	if err != nil {
		return nil, err
	}

	// Private messages & channels have not been implemented yet
	message.Target = "#osu"
	message.Sender = ""

	return message, nil
}

func (client *B282) ReadFrameBundle(reader io.Reader) (*chio.ReplayFrameBundle, error) {
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

	return &chio.ReplayFrameBundle{Frames: frames, Action: action}, nil
}

func (client *B282) ReadReplayFrame(reader io.Reader) (*chio.ReplayFrame, error) {
	var err error
	errors := internal.NewErrorCollection()
	frame := &chio.ReplayFrame{}
	mouseLeft, err := internal.ReadBoolean(reader)
	errors.Add(err)
	mouseRight, err := internal.ReadBoolean(reader)
	errors.Add(err)
	frame.MouseX, err = internal.ReadFloat32(reader)
	errors.Add(err)
	frame.MouseY, err = internal.ReadFloat32(reader)
	errors.Add(err)
	frame.Time, err = internal.ReadInt32(reader)
	errors.Add(err)

	frame.ButtonState = 0

	if mouseLeft {
		frame.ButtonState |= chio.ButtonStateLeft1
	}
	if mouseRight {
		frame.ButtonState |= chio.ButtonStateRight1
	}

	return frame, errors.Next()
}

func NewB282() *B282 {
	client := &B282{
		SlotSize:    8,
		ProtocolVer: 0,
		Readers:     make(chio.ReaderRegistry),
	}

	client.Readers[chio.OsuSendUserStatus] = internal.ReaderReadStatus()
	client.Readers[chio.OsuSendIrcMessage] = internal.ReaderReadMessage()
	client.Readers[chio.OsuStartSpectating] = internal.ReaderReadBanchoInt()
	client.Readers[chio.OsuSpectateFrames] = internal.ReaderReadFrameBundle()
	client.Readers[chio.OsuErrorReport] = internal.ReaderReadBanchoString()

	client.SupportedPacketIds = []uint16{
		chio.OsuSendUserStatus,
		chio.OsuSendIrcMessage,
		chio.OsuExit,
		chio.OsuRequestStatusUpdate,
		chio.OsuPong,
		chio.BanchoLoginReply,
		chio.BanchoCommandError,
		chio.BanchoSendMessage,
		chio.BanchoPing,
		chio.BanchoHandleIrcChangeUsername,
		chio.BanchoHandleIrcQuit,
		chio.BanchoHandleOsuUpdate,
		chio.BanchoHandleOsuQuit,
		chio.BanchoSpectatorJoined,
		chio.BanchoSpectatorLeft,
		chio.BanchoSpectateFrames,
		chio.OsuStartSpectating,
		chio.OsuStopSpectating,
		chio.OsuSpectateFrames,
		chio.BanchoVersionUpdate,
		chio.OsuErrorReport,
		chio.OsuCantSpectate,
		chio.BanchoSpectatorCantSpectate,
	}

	return client
}

func init() {
	client := NewB282()
	chio.RegisterClient(282, client)
	chio.RegisterClient(290, client)
}

/* Unsupported Packets */

func (client *B282) WriteGetAttention(stream io.Writer) error                        { return nil }
func (client *B282) WriteAnnouncement(stream io.Writer, message string) error        { return nil }
func (client *B282) WriteMatchUpdate(stream io.Writer, match chio.Match) error       { return nil }
func (client *B282) WriteMatchNew(stream io.Writer, match chio.Match) error          { return nil }
func (client *B282) WriteMatchDisband(stream io.Writer, matchId int32) error         { return nil }
func (client *B282) WriteLobbyJoin(stream io.Writer, userId int32) error             { return nil }
func (client *B282) WriteLobbyPart(stream io.Writer, userId int32) error             { return nil }
func (client *B282) WriteMatchJoinSuccess(stream io.Writer, match chio.Match) error  { return nil }
func (client *B282) WriteMatchJoinFail(stream io.Writer) error                       { return nil }
func (client *B282) WriteFellowSpectatorJoined(stream io.Writer, userId int32) error { return nil }
func (client *B282) WriteFellowSpectatorLeft(stream io.Writer, userId int32) error   { return nil }
func (client *B282) WriteMatchStart(stream io.Writer, match chio.Match) error        { return nil }
func (client *B282) WriteMatchScoreUpdate(stream io.Writer, frame chio.ScoreFrame) error {
	return nil
}
func (client *B282) WriteMatchTransferHost(stream io.Writer) error                  { return nil }
func (client *B282) WriteMatchAllPlayersLoaded(stream io.Writer) error              { return nil }
func (client *B282) WriteMatchPlayerFailed(stream io.Writer, slotId uint32) error   { return nil }
func (client *B282) WriteMatchComplete(stream io.Writer) error                      { return nil }
func (client *B282) WriteMatchSkip(stream io.Writer) error                          { return nil }
func (client *B282) WriteUnauthorized(stream io.Writer) error                       { return nil }
func (client *B282) WriteChannelJoinSuccess(stream io.Writer, channel string) error { return nil }
func (client *B282) WriteChannelRevoked(stream io.Writer, channel string) error     { return nil }
func (client *B282) WriteChannelAvailable(stream io.Writer, channel chio.Channel) error {
	return nil
}
func (client *B282) WriteChannelAvailableAutojoin(stream io.Writer, channel chio.Channel) error {
	return nil
}
func (client *B282) WriteBeatmapInfoReply(stream io.Writer, reply chio.BeatmapInfoReply) error {
	return nil
}
func (client *B282) WriteLoginPermissions(stream io.Writer, permissions uint32) error { return nil }
func (client *B282) WriteFriendsList(stream io.Writer, userIds []int32) error         { return nil }
func (client *B282) WriteProtocolNegotiation(stream io.Writer, version int32) error   { return nil }
func (client *B282) WriteTitleUpdate(stream io.Writer, update chio.TitleUpdate) error { return nil }
func (client *B282) WriteMonitor(stream io.Writer) error                              { return nil }
func (client *B282) WriteMatchPlayerSkipped(stream io.Writer, slotId int32) error     { return nil }
func (client *B282) WriteRestart(stream io.Writer, retryMs int32) error               { return nil }
func (client *B282) WriteInvite(stream io.Writer, message chio.Message) error         { return nil }
func (client *B282) WriteChannelInfoComplete(stream io.Writer) error                  { return nil }
func (client *B282) WriteMatchChangePassword(stream io.Writer, password string) error { return nil }
func (client *B282) WriteSilenceInfo(stream io.Writer, timeRemaining int32) error     { return nil }
func (client *B282) WriteUserSilenced(stream io.Writer, userId uint32) error          { return nil }
func (client *B282) WriteUserDMsBlocked(stream io.Writer, targetName string) error    { return nil }
func (client *B282) WriteTargetIsSilenced(stream io.Writer, targetName string) error  { return nil }
func (client *B282) WriteVersionUpdateForced(stream io.Writer) error                  { return nil }
func (client *B282) WriteSwitchServer(stream io.Writer, target int32) error           { return nil }
func (client *B282) WriteAccountRestricted(stream io.Writer) error                    { return nil }
func (client *B282) WriteRTX(stream io.Writer, message string) error                  { return nil }
func (client *B282) WriteMatchAbort(stream io.Writer) error                           { return nil }
func (client *B282) WriteSwitchTournamentServer(stream io.Writer, ip string) error    { return nil }
