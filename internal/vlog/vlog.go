package vlog

import (
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"sync"

	"github.com/edsrzf/mmap-go"
)

const MaxFileSize = 1 << 20

type logFile struct {
	file *os.File
	data mmap.MMap
}

type VLog struct {
	dir string

	mu sync.RWMutex

	activeFid uint32
	offset    int64

	files map[uint32]*logFile
}

type ValuePointer struct {
	Fid    uint32
	Offset int64
	Len    int64
}

type ReadResult struct {
	Ptr   ValuePointer
	Value []byte
	Err   error
}

func OpenVLog(dir string) (*VLog, error) {

	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("open vlog: %w", err)
	}

	v := &VLog{
		dir:   dir,
		files: make(map[uint32]*logFile),
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("open vlog: %w", err)
	}

	var activeFid uint32 = 0

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		fid64, err := strconv.ParseUint(entry.Name(), 10, 32)
		if err != nil {
			continue
		}

		fid := uint32(fid64)

		file, err := os.OpenFile(
			filepath.Join(dir, strconv.Itoa(int(fid))),
			os.O_RDWR,
			0644,
		)
		if err != nil {
			return nil, fmt.Errorf("open vlog: %w", err)
		}

		v.files[fid] = &logFile{
			file: file,
		}

		if fid > activeFid {
			activeFid = fid
		}

	}

	if activeFid == 0 {
		activeFid = 1

		file, err := os.OpenFile(
			filepath.Join(dir, "1"),
			os.O_CREATE|os.O_RDWR|os.O_APPEND,
			0644,
		)
		if err != nil {
			return nil, fmt.Errorf("open vlog: %w", err)
		}

		v.files[1] = &logFile{
			file: file,
		}
	}

	active := v.files[activeFid]

	stat, err := active.file.Stat()
	if err != nil {
		return nil, fmt.Errorf("open vlog: %w", err)
	}

	isNew := stat.Size() == 0

	if stat.Size() < MaxFileSize {
		if err := active.file.Truncate(MaxFileSize); err != nil {
			return nil, fmt.Errorf("truncate active vlog: %w", err)
		}
	}

	data, err := mmap.MapRegion(active.file, MaxFileSize, mmap.RDWR, 0, 0)

	active.data = data

	v.activeFid = activeFid

	if isNew {
		v.offset = 0
	} else {
		offset, err := findOffset(data)
		if err != nil {
			data.Unmap()
			active.file.Close()
			return nil, fmt.Errorf("recover vlog offset: %w", err)
		}

		v.offset = offset
	}

	return v, nil
}

func (v *VLog) Close() error {
	v.mu.Lock()
	defer v.mu.Unlock()

	var firstErr error

	for fid, lf := range v.files {
		if lf.data != nil {
			if err := lf.data.Unmap(); err != nil && firstErr == nil {
				firstErr = fmt.Errorf("unmap fid %d: %w", fid, err)
			}
		}

		if err := lf.file.Close(); err != nil && firstErr == nil {
			firstErr = fmt.Errorf("close fid %d: %w", fid, err)
		}
	}

	return firstErr
}

func (v *VLog) Sync() error {
	v.mu.RLock()
	defer v.mu.RUnlock()

	return v.syncLocked()
}

func (v *VLog) syncLocked() error {
	active := v.files[v.activeFid]

	if active.data == nil {
		return fmt.Errorf("sync vlog: active mmap is nil")
	}

	if err := active.data.Flush(); err != nil {
		return fmt.Errorf("sync vlog: mmap flush: %w", err)
	}

	if err := active.file.Sync(); err != nil {
		return fmt.Errorf("sync vlog: file sync: %w", err)
	}

	return nil
}

func findOffset(data []byte) (int64, error) {
	var offset int64

	for offset+8 <= int64(len(data)) {
		header := data[offset : offset+8]

		keyLen := binary.BigEndian.Uint32(header[0:4])
		valueLen := binary.BigEndian.Uint32(header[4:8])

		if keyLen == 0 && valueLen == 0 {
			return offset, nil
		}

		entrySize := int64(8) +
			int64(keyLen) +
			int64(valueLen)

		if entrySize < 8 {
			return 0, fmt.Errorf("invalid entry size at offset %d", offset)
		}

		if offset+entrySize > int64(len(data)) {
			return offset, nil
		}

		offset += entrySize
	}

	return offset, nil
}
