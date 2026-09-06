package vlog

import (
	"fmt"
	"io"
)

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
