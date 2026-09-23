package vlog

import (
	"encoding/binary"
	"fmt"
)

type Entry struct {
	Key      []byte
	Value    []byte
	Sequence uint64
	VType    uint8
}

func EncodeEntry(entry Entry, buf []byte) int {
	keyLen := len(entry.Key)
	valueLen := len(entry.Value)

	n := 8 + keyLen + valueLen

	binary.BigEndian.PutUint32(buf[0:4], uint32(keyLen))
	binary.BigEndian.PutUint32(buf[4:8], uint32(valueLen))

	copy(buf[8:8+keyLen], entry.Key)
	copy(buf[8+keyLen:], entry.Value)

	binary.BigEndian.PutUint64(buf[n:n+8], entry.Sequence)
	buf[n+8] = entry.VType

	return n + 9
}

func DecodeEntry(buf []byte) (Entry, error) {
	var entry Entry

	if len(buf) < 17 {
		return entry, fmt.Errorf("decode entry: buffer too small")
	}

	keyLen := binary.BigEndian.Uint32(buf[0:4])
	valueLen := binary.BigEndian.Uint32(buf[4:8])

	n := 8 + int(keyLen) + int(valueLen)
	totalLen := n + 9

	if len(buf) < totalLen {
		return entry, fmt.Errorf("decode entry: incomplete entry")
	}

	entry.Key = make([]byte, keyLen)
	entry.Value = make([]byte, valueLen)

	copy(entry.Key, buf[8:8+keyLen])
	copy(entry.Value, buf[8+keyLen:n])

	entry.Sequence = binary.BigEndian.Uint64(buf[n : n+8])
	entry.VType = buf[n+8]

	return entry, nil
}
