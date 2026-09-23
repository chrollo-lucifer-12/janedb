package vlog

import (
	"io"
)

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
