package keys

import (
	"bytes"
	"encoding/binary"
)

func InternalKey(key []byte, sequence uint64, vType uint8) []byte {
	tag := (sequence << 8) | uint64(vType)

	buf := make([]byte, len(key)+8)
	copy(buf, key)

	binary.LittleEndian.PutUint64(buf[len(key):], tag)

	return buf
}

func CompareInternalKey(a, b []byte) int {
	aUser := a[:len(a)-8]
	bUser := b[:len(b)-8]

	if cmp := bytes.Compare(aUser, bUser); cmp != 0 {
		return cmp
	}

	aTag := binary.LittleEndian.Uint64(a[len(a)-8:])
	bTag := binary.LittleEndian.Uint64(b[len(b)-8:])

	if aTag > bTag {
		return -1
	}
	if aTag < bTag {
		return 1
	}

	return 0
}
