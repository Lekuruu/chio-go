package internal

import (
	"fmt"
	"io"

	chio "github.com/Lekuruu/chio-go"
)

// NOTE: Packet readers are registered once in the base client version.
// 		 At runtime, we dispatch through interfaces to the concrete client instance.
// 		 This means derived client versions do NOT need to re-register packet readers
// 		 just to change behavior. If a newer version overrides a read method, the
// 	     dispatch automatically calls the overridden implementation.

// Reader interfaces
type (
	StatusReader interface {
		ReadStatus(io.Reader) (*chio.UserStatus, error)
	}
	MessageReader interface {
		ReadMessage(io.Reader) (*chio.Message, error)
	}
	PrivateMessageReader interface {
		ReadPrivateMessage(io.Reader) (*chio.Message, error)
	}
	FrameBundleReader interface {
		ReadFrameBundle(io.Reader) (*chio.ReplayFrameBundle, error)
	}
	MatchReader interface {
		ReadMatch(io.Reader) (*chio.Match, error)
	}
	MatchJoinReader interface {
		ReadMatchJoin(io.Reader) (*chio.MatchJoin, error)
	}
	MatchChangeSlotReader interface {
		ReadMatchChangeSlot(io.Reader) (int32, error)
	}
	MatchLockReader interface {
		ReadMatchLock(io.Reader) (int32, error)
	}
)

// dispatchReader creates a PacketReader that delegates to the client's method.
func dispatchReader[T any](name string, extract func(chio.BanchoIO) (func(io.Reader) (T, error), bool)) chio.PacketReader {
	return func(client chio.BanchoIO, r io.Reader) (any, error) {
		if fn, ok := extract(client); ok {
			return fn(r)
		}
		return nil, fmt.Errorf("client does not implement %s", name)
	}
}

func ReaderReadStatus() chio.PacketReader {
	return dispatchReader("ReadStatus", func(c chio.BanchoIO) (func(io.Reader) (*chio.UserStatus, error), bool) {
		if h, ok := c.(StatusReader); ok {
			return h.ReadStatus, true
		}
		return nil, false
	})
}

func ReaderReadMessage() chio.PacketReader {
	return dispatchReader("ReadMessage", func(c chio.BanchoIO) (func(io.Reader) (*chio.Message, error), bool) {
		if h, ok := c.(MessageReader); ok {
			return h.ReadMessage, true
		}
		return nil, false
	})
}

func ReaderReadFrameBundle() chio.PacketReader {
	return dispatchReader("ReadFrameBundle", func(c chio.BanchoIO) (func(io.Reader) (*chio.ReplayFrameBundle, error), bool) {
		if h, ok := c.(FrameBundleReader); ok {
			return h.ReadFrameBundle, true
		}
		return nil, false
	})
}

func ReaderReadPrivateMessage() chio.PacketReader {
	return dispatchReader("ReadPrivateMessage", func(c chio.BanchoIO) (func(io.Reader) (*chio.Message, error), bool) {
		if h, ok := c.(PrivateMessageReader); ok {
			return h.ReadPrivateMessage, true
		}
		return nil, false
	})
}

func ReaderReadMatch() chio.PacketReader {
	return dispatchReader("ReadMatch", func(c chio.BanchoIO) (func(io.Reader) (*chio.Match, error), bool) {
		if h, ok := c.(MatchReader); ok {
			return h.ReadMatch, true
		}
		return nil, false
	})
}

func ReaderReadMatchJoin() chio.PacketReader {
	return dispatchReader("ReadMatchJoin", func(c chio.BanchoIO) (func(io.Reader) (*chio.MatchJoin, error), bool) {
		if h, ok := c.(MatchJoinReader); ok {
			return h.ReadMatchJoin, true
		}
		return nil, false
	})
}

func ReaderReadMatchChangeSlot() chio.PacketReader {
	return dispatchReader("ReadMatchChangeSlot", func(c chio.BanchoIO) (func(io.Reader) (int32, error), bool) {
		if h, ok := c.(MatchChangeSlotReader); ok {
			return h.ReadMatchChangeSlot, true
		}
		return nil, false
	})
}

func ReaderReadMatchLock() chio.PacketReader {
	return dispatchReader("ReadMatchLock", func(c chio.BanchoIO) (func(io.Reader) (int32, error), bool) {
		if h, ok := c.(MatchLockReader); ok {
			return h.ReadMatchLock, true
		}
		return nil, false
	})
}

func ReaderReadEmpty() chio.PacketReader {
	return func(_ chio.BanchoIO, r io.Reader) (any, error) {
		return nil, nil
	}
}

// Simple readers for primitive types, e.g. bInt or bString

func ReaderReadBanchoInt() chio.PacketReader {
	return func(_ chio.BanchoIO, r io.Reader) (any, error) {
		return ReadInt32(r)
	}
}

func ReaderReadBanchoString() chio.PacketReader {
	return func(_ chio.BanchoIO, r io.Reader) (any, error) {
		return ReadString(r)
	}
}
