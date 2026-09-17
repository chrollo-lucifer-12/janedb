package vlog

import (
	"io"
)

func (v *VLog) Recover() map[string]ValuePointer {

	values := make(map[string]ValuePointer)

	start := int64(0)

	for start < int64(len(v.data)) {

		entry, err := DecodeEntry(v.data[start:])
		if err != nil {
			break
		}

		if entry.Value == nil {
			delete(values, string(entry.Key))
		} else {
			values[string(entry.Key)] = ValuePointer{
				Offset: start + 8 + int64(len(entry.Key)),
				Len:    int64(len(entry.Value)),
			}
		}

		start += 8 + int64(len(entry.Key)) + int64(len(entry.Value))
	}

	return values
}

func (v *VLog) ReadValue(ptr ValuePointer, buf []byte) error {

	start := ptr.Offset
	end := start + ptr.Len

	if end > int64(len(v.data)) {
		return io.ErrUnexpectedEOF
	}

	if int64(len(buf)) < ptr.Len {
		return io.ErrShortBuffer
	}

	copy(buf, v.data[start:end])

	return nil
}
