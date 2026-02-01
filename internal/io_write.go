package internal

import (
	"encoding/binary"
	"errors"
	"io"

	"github.com/bnch/uleb128"
)

func WriteUint64(w io.Writer, v uint64) error {
	return binary.Write(w, binary.LittleEndian, v)
}

func WriteInt64(w io.Writer, v int64) error {
	return WriteUint64(w, uint64(v))
}

func WriteUint32(w io.Writer, v uint32) error {
	return binary.Write(w, binary.LittleEndian, v)
}

func WriteInt32(w io.Writer, v int32) error {
	return WriteUint32(w, uint32(v))
}

func WriteUint16(w io.Writer, v uint16) error {
	return binary.Write(w, binary.LittleEndian, v)
}

func WriteInt16(w io.Writer, v int16) error {
	return WriteUint16(w, uint16(v))
}

func WriteUint8(w io.Writer, v uint8) error {
	return binary.Write(w, binary.LittleEndian, v)
}

func WriteInt8(w io.Writer, v int8) error {
	return WriteUint8(w, uint8(v))
}

func WriteBoolean(w io.Writer, v bool) error {
	return binary.Write(w, binary.LittleEndian, v)
}

func WriteFloat32(w io.Writer, v float32) error {
	return binary.Write(w, binary.LittleEndian, v)
}

func WriteFloat64(w io.Writer, v float64) error {
	return binary.Write(w, binary.LittleEndian, v)
}

func WriteIntList16(w io.Writer, v []int32) error {
	if err := WriteUint16(w, uint16(len(v))); err != nil {
		return err
	}
	for _, i := range v {
		if err := WriteInt32(w, i); err != nil {
			return err
		}
	}
	return nil
}

func WriteIntList32(w io.Writer, v []int32) error {
	if err := WriteUint32(w, uint32(len(v))); err != nil {
		return err
	}
	for _, i := range v {
		if err := WriteInt32(w, i); err != nil {
			return err
		}
	}
	return nil
}

func WriteBoolList(w io.Writer, bools []bool) error {
	if len(bools) < 8 {
		return errors.New("bool list must have at least 8 elements")
	}

	var result byte
	for i := 7; i >= 0; i-- {
		if bools[i] {
			result |= 1
		}
		if i > 0 {
			result <<= 1
		}
	}
	return WriteUint8(w, result)
}

func WriteString(w io.Writer, v string) error {
	if v == "" {
		binary.Write(w, binary.LittleEndian, uint8(0x00))
		return nil
	}

	if err := binary.Write(w, binary.LittleEndian, uint8(0x0b)); err != nil {
		return err
	}

	w.Write(uleb128.Marshal(len(v)))
	w.Write([]byte(v))
	return nil
}
