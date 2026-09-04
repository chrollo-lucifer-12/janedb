package vlog

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

func (v *VLog) Rotate() error {
	if err := v.Sync(); err != nil {
		return fmt.Errorf("rotate vlog: %w", err)
	}

	newFid := v.activeFid + 1
	newVLogFilename := strconv.Itoa(int(newFid))
	newVLogPath := filepath.Join(v.dir, newVLogFilename)

	file, err := os.OpenFile(newVLogPath, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return fmt.Errorf("rotate vlog: %w", err)
	}

	v.activeFid = newFid
	v.active = file
	v.files[v.activeFid] = file
	v.writer = bufio.NewWriterSize(file, 64*1024)
	v.offset = 0

	return nil
}

func (v *VLog) Append(entry Entry) (ValuePointer, error) {

	if v.offset >= MaxFileSize {
		if err := v.Rotate(); err != nil {
			return ValuePointer{}, fmt.Errorf("append entry: %w", err)
		}
	}

	v.mu.Lock()

	if err := EncodeEntry(entry, v.writer); err != nil {
		return ValuePointer{}, fmt.Errorf("append entry: %w", err)
	}

	v.mu.Unlock()

	valueOffset := v.offset + 8 + int64(len(entry.key))
	valueLen := int64(len(entry.value))

	ptr := ValuePointer{
		Fid:    v.activeFid,
		Offset: valueOffset,
		Len:    valueLen,
	}

	v.offset += 8 + int64(len(entry.key)) + valueLen

	return ptr, nil
}
