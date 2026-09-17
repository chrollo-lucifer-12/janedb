package vlog

import (
	"encoding/binary"
	"fmt"
)

type Entry struct {
	Key   []byte
	Value []byte
}

func EncodeEntry(entry Entry, buf []byte) int {
	keyLen := len(entry.Key)
	valueLen := len(entry.Value)

	n := 8 + keyLen + valueLen

	binary.BigEndian.PutUint32(buf[0:4], uint32(keyLen))
	binary.BigEndian.PutUint32(buf[4:8], uint32(valueLen))

	copy(buf[8:8+keyLen], entry.Key)
	copy(buf[8+keyLen:], entry.Value)

	return n
}

func DecodeEntry(buf []byte) (Entry, error) {

	var entry Entry

	if len(buf) < 8 {
		return entry, fmt.Errorf("decode entry: buffer too small")
	}

	keyLen := binary.BigEndian.Uint32(buf[0:4])
	valueLen := binary.BigEndian.Uint32(buf[4:8])

	totalLen := 8 + int(keyLen) + int(valueLen)

	if len(buf) < totalLen {
		return entry, fmt.Errorf("decode entry: incomplete entry")
	}

	entry.Key = make([]byte, keyLen)
	entry.Value = make([]byte, valueLen)

	copy(entry.Key, buf[8:8+keyLen])
	copy(entry.Value, buf[8+keyLen:totalLen])

	return entry, nil
}
