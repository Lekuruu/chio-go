package internal

import (
	"bytes"
	"compress/gzip"
	"io"
)

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
