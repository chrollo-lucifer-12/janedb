package vlog

import (
	"encoding/binary"
	"fmt"
	"io"
)

type Entry struct {
	key   []byte
	value []byte
}

func NewEntry(key, value []byte) Entry {
	return Entry{key: key, value: value}
}

func EncodeEntry(entry Entry) ([]byte, error) {
	keyLen := len(entry.key)
	valueLen := len(entry.value)

	buf := make([]byte, 8+keyLen+valueLen)

	binary.BigEndian.PutUint32(buf[0:4], uint32(keyLen))
	binary.BigEndian.PutUint32(buf[4:8], uint32(valueLen))

	copy(buf[8:8+keyLen], entry.key)
	copy(buf[8+keyLen:], entry.value)

	return buf, nil
}

func DecodeEntry(r io.Reader) (Entry, error) {

	var entry Entry

	var header [8]byte

	if _, err := io.ReadFull(r, header[:]); err != nil {
		return entry, fmt.Errorf("decode entry: read header: %w", err)
	}

	keyLen := binary.BigEndian.Uint32(header[0:4])
	valueLen := binary.BigEndian.Uint32(header[4:8])

	entry.key = make([]byte, keyLen)
	entry.value = make([]byte, valueLen)

	if _, err := io.ReadFull(r, entry.key); err != nil {
		return Entry{}, fmt.Errorf("decode entry: read key: %w", err)
	}

	if _, err := io.ReadFull(r, entry.value); err != nil {
		return Entry{}, fmt.Errorf("decode entry: read value: %w", err)
	}

	return entry, nil
}
