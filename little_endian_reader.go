package zipstream

import (
	"encoding/binary"
	"io"
)

// LittleEndianReader helps to read data in LittleEndian. Zip headers are written in LittleEndian.
type LittleEndianReader struct {
	io.Reader
}

func (l *LittleEndianReader) uint16() uint16 {
	buff := make([]byte, 2)
	l.Read(buff)
	return binary.LittleEndian.Uint16(buff)
}

func (l *LittleEndianReader) uint32() uint32 {
	buff := make([]byte, 4)
	l.Read(buff)
	return binary.LittleEndian.Uint32(buff)
}

