package internal

import (
	"encoding/binary"
	"errors"
	"io"

	"github.com/bnch/uleb128"
)

func ReadUint64(r io.Reader) (v uint64, err error) {
	err = binary.Read(r, binary.LittleEndian, &v)
	return v, err
}

func ReadInt64(r io.Reader) (v int64, err error) {
	uv, err := ReadUint64(r)
	return int64(uv), err
}

func ReadUint32(r io.Reader) (v uint32, err error) {
	err = binary.Read(r, binary.LittleEndian, &v)
	return v, err
}

func ReadInt32(r io.Reader) (v int32, err error) {
	uv, err := ReadUint32(r)
	return int32(uv), err
}

func ReadUint16(r io.Reader) (v uint16, err error) {
	err = binary.Read(r, binary.LittleEndian, &v)
	return v, err
}

func ReadInt16(r io.Reader) (v int16, err error) {
	uv, err := ReadUint16(r)
	return int16(uv), err
}

func ReadUint8(r io.Reader) (v uint8, err error) {
	err = binary.Read(r, binary.LittleEndian, &v)
	return v, err
}

func ReadInt8(r io.Reader) (v int8, err error) {
	uv, err := ReadUint8(r)
	return int8(uv), err
}

func ReadBoolean(r io.Reader) (v bool, err error) {
	err = binary.Read(r, binary.LittleEndian, &v)
	return v, err
}

func ReadFloat32(r io.Reader) (v float32, err error) {
	err = binary.Read(r, binary.LittleEndian, &v)
	return v, err
}

func ReadFloat64(r io.Reader) (v float64, err error) {
	err = binary.Read(r, binary.LittleEndian, &v)
	return v, err
}

func ReadIntList16(r io.Reader) (v []int32, err error) {
	l, err := ReadUint16(r)
	if err != nil {
		return nil, err
	}

	v = make([]int32, l)
	for i := uint16(0); i < l; i++ {
		v[i], err = ReadInt32(r)
		if err != nil {
			return nil, err
		}
	}

	return v, nil
}

func ReadIntList32(r io.Reader) (v []int32, err error) {
	l, err := ReadUint32(r)
	if err != nil {
		return nil, err
	}

	v = make([]int32, l)
	for i := uint32(0); i < l; i++ {
		v[i], err = ReadInt32(r)
		if err != nil {
			return nil, err
		}
	}

	return v, nil
}

func ReadBoolList(r io.Reader, size int) ([]bool, error) {
	input, err := ReadUint8(r)
	if err != nil {
		return nil, err
	}

	bools := make([]bool, size)
	for i := 0; i < size; i++ {
		bools[i] = ((input >> i) & 1) > 0
	}
	return bools, nil
}

func ReadString(r io.Reader) (v string, err error) {
	var b uint8
	err = binary.Read(r, binary.LittleEndian, &b)
	if err != nil {
		return "", err
	}

	if b == 0x00 {
		return "", nil
	}

	if b != 0x0b {
		return "", errors.New("invalid string type")
	}

	l := uleb128.UnmarshalReader(r)
	buf := make([]byte, l)
	_, err = r.Read(buf)
	if err != nil {
		return "", err
	}

	return string(buf), nil
}
