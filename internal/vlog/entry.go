package vlog

import (
	"encoding/binary"
	"fmt"
)

type Entry struct {
	key   []byte
	value []byte
}

func NewEntry(key, value []byte) Entry {
	return Entry{key: key, value: value}
}

func EncodeEntry(entry Entry, buf []byte) int {
	keyLen := len(entry.key)
	valueLen := len(entry.value)

	n := 8 + keyLen + valueLen

	binary.BigEndian.PutUint32(buf[0:4], uint32(keyLen))
	binary.BigEndian.PutUint32(buf[4:8], uint32(valueLen))

	copy(buf[8:8+keyLen], entry.key)
	copy(buf[8+keyLen:], entry.value)

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

	entry.key = make([]byte, keyLen)
	entry.value = make([]byte, valueLen)

	copy(entry.key, buf[8:8+keyLen])
	copy(entry.value, buf[8+keyLen:totalLen])

	return entry, nil
}
