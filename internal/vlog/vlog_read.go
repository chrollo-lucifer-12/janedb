package vlog

import (
	"fmt"
	"io"
)

func (v *VLog) Recover() map[string]ValuePointer {

	values := make(map[string]ValuePointer)

	for fid, file := range v.files {

		start := int64(0)

		for start < v.offset {

			entry, err := DecodeEntry(file.data[start:])
			if err != nil {
				break
			}

			if entry.value == nil {
				delete(values, string(entry.key))
			} else {
				values[string(entry.key)] = ValuePointer{
					Fid:    fid,
					Offset: start + 8 + int64(len(entry.key)),
					Len:    int64(len(entry.value)),
				}
			}

			start += 8 + int64(len(entry.key)) + int64(len(entry.value))
		}
	}

	return values
}

func (v *VLog) ReadValue(ptr ValuePointer, buf []byte) error {
	v.mu.RLock()
	defer v.mu.RUnlock()

	file, ok := v.files[ptr.Fid]
	if !ok {
		return fmt.Errorf("read vlog: file not found for fid %d", ptr.Fid)
	}

	start := ptr.Offset
	end := start + ptr.Len

	if end > int64(len(file.data)) {
		return io.ErrUnexpectedEOF
	}

	if int64(len(buf)) < ptr.Len {
		return io.ErrShortBuffer
	}

	copy(buf, file.data[start:end])

	return nil
}
