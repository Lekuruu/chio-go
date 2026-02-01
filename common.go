package chio

import (
	"bytes"
	"compress/gzip"
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

func CompressData(data []byte) []byte {
	if len(data) == 0 {
		return []byte{}
	}
	zb := new(bytes.Buffer)
	zw := gzip.NewWriter(zb)
	zw.Write(data)
	zw.Close()
	return zb.Bytes()
}

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

func ReadBoolList(r io.Reader) ([]bool, error) {
	input, err := ReadUint8(r)
	if err != nil {
		return nil, err
	}

	bools := make([]bool, 8)
	for i := 0; i < 8; i++ {
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

func DecompressData(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return []byte{}, nil
	}
	dst := bytes.NewBuffer([]byte{})
	zr, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	io.Copy(dst, zr)
	zr.Close()
	return dst.Bytes(), nil
}
