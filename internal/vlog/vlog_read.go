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

			if entry.Value == nil {
				delete(values, string(entry.Key))
			} else {
				values[string(entry.Key)] = ValuePointer{
					Fid:    fid,
					Offset: start + 8 + int64(len(entry.Key)),
					Len:    int64(len(entry.Value)),
				}
			}

			start += 8 + int64(len(entry.Key)) + int64(len(entry.Value))
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
