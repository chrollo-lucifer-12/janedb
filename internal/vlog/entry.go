package vlog

import (
	"encoding/binary"
	"fmt"
)

type Entry struct {
	key   []byte
	value []byte
}

func EncodeEntry(entry Entry, buf []byte) error {
	keyLen := len(entry.key)
	valueLen := len(entry.value)

	if len(buf) < 8+keyLen+valueLen {
		return fmt.Errorf(
			"encode entry: buffer too small: got %d, need %d",
			len(buf),
			8+keyLen+valueLen,
		)
	}

	binary.BigEndian.PutUint32(buf, uint32(keyLen))
	binary.BigEndian.PutUint32(buf[4:8], uint32(valueLen))
	copy(buf[8:8+keyLen], entry.key)
	copy(buf[8+keyLen:8+keyLen+valueLen], entry.value)

	return nil
}

func DecodeEntry(buf []byte) (Entry, error) {

	var entry Entry

	if len(buf) < 8 {
		return entry, fmt.Errorf("decode entry: incomplete header")
	}

	keyLen := binary.BigEndian.Uint32(buf)
	valueLen := binary.BigEndian.Uint32(buf[4:])

	total := 8 + int(keyLen) + int(valueLen)

	if len(buf) < total {
		return entry, fmt.Errorf("decode entry: incomeplete entry: got %d bytes need %d", len(buf), total)
	}

	entry.key = make([]byte, keyLen)
	entry.value = make([]byte, valueLen)

	copy(entry.key, buf[8:8+int(keyLen)])
	copy(entry.value, buf[8+int(keyLen):8+int(keyLen)+int(valueLen)])

	return entry, nil
}
