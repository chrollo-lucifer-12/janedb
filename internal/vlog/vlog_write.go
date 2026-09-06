package vlog

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/edsrzf/mmap-go"
)

func (v *VLog) Append(entry Entry) (ValuePointer, error) {

	v.mu.Lock()
	defer v.mu.Unlock()

	entryLen := int64(8 + len(entry.key) + len(entry.value))

	if v.offset+entryLen > MaxFileSize {
		if err := v.rotateLocked(); err != nil {
			return ValuePointer{}, fmt.Errorf("append entry: %w", err)
		}
	}

	active := v.files[v.activeFid]

	start := v.offset

	EncodeEntry(entry, active.data[start:])

	ptr := ValuePointer{
		Fid:    v.activeFid,
		Offset: start + 8 + int64(len(entry.key)),
		Len:    int64(len(entry.value)),
	}

	v.offset += entryLen

	return ptr, nil
}

func (v *VLog) rotate() error {
	v.mu.Lock()
	defer v.mu.Unlock()

	return v.rotateLocked()
}

func (v *VLog) rotateLocked() error {
	if err := v.syncLocked(); err != nil {
		return fmt.Errorf("rotate vlog: %w", err)
	}

	newFid := v.activeFid + 1
	newVLogFilename := strconv.Itoa(int(newFid))
	newVLogPath := filepath.Join(v.dir, newVLogFilename)

	file, err := os.OpenFile(newVLogPath, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return fmt.Errorf("rotate vlog: %w", err)
	}

	if err := file.Truncate(MaxFileSize); err != nil {
		file.Close()
		return fmt.Errorf("rotate vlog: truncate: %w", err)
	}

	data, err := mmap.MapRegion(
		file,
		MaxFileSize,
		mmap.RDWR,
		0,
		0,
	)
	if err != nil {
		file.Close()
		return fmt.Errorf("rotate vlog: mmap: %w", err)
	}

	v.activeFid = newFid
	v.files[v.activeFid] = &logFile{file: file, data: data}
	v.offset = 0

	return nil
}
