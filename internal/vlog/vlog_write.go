package vlog

import (
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
	v.files[v.activeFid] = &logFile{file: file}
	v.offset = 0

	return nil
}

func (v *VLog) Append(entry Entry) (ValuePointer, error) {

	v.mu.Lock()
	defer v.mu.Unlock()

	if v.offset >= MaxFileSize {
		if err := v.Rotate(); err != nil {
			return ValuePointer{}, fmt.Errorf("append entry: %w", err)
		}
	}

	buf, err := EncodeEntry(entry)
	if err != nil {
		return ValuePointer{}, err
	}

	entryLen := int64(len(buf))

	active := v.files[v.activeFid]

	copy(active.data[v.offset:v.offset+entryLen], buf)

	valueOffset := v.offset + 8 + int64(len(entry.key))

	ptr := ValuePointer{
		Fid:    v.activeFid,
		Offset: valueOffset,
		Len:    int64(len(entry.value)),
	}

	v.offset += entryLen

	return ptr, nil
}
