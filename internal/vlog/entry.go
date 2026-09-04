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

func EncodeEntry(entry Entry, w io.Writer) error {
	keyLen := len(entry.key)
	valueLen := len(entry.value)

	var header [8]byte

	binary.BigEndian.PutUint32(header[0:4], uint32(keyLen))
	binary.BigEndian.PutUint32(header[4:8], uint32(valueLen))

	if _, err := w.Write(header[:]); err != nil {
		return fmt.Errorf("encode entry: write header: %w", err)
	}

	if _, err := w.Write(entry.key); err != nil {
		return fmt.Errorf("encode entry: write key: %w", err)
	}

	if _, err := w.Write(entry.value); err != nil {
		return fmt.Errorf("encode entry: write value: %w", err)
	}

	return nil
}

func DecodeEntry(r io.Reader) (Entry, error) {

	var entry Entry

	var header [8]byte

	if _, err := io.ReadFull(r, header[:]); err != nil {
		return entry, fmt.Errorf("decode entry: read header: %w", err)
	}

	keyLen := binary.BigEndian.Uint32(header[0:4])
	valueLen := binary.BigEndian.Uint32(header[4:8])

	total := 8 + int(keyLen) + int(valueLen)

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
