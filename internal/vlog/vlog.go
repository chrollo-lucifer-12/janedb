package vlog

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
)

const MaxFileSize = 1 << 20

type VLog struct {
	dir string

	active *os.File
	writer *bufio.Writer

	activeFid uint32
	offset    int64

	files map[uint32]*os.File
}

type ValuePointer struct {
	Fid    uint32
	Offset int64
	Len    int64
}

func OpenVLog(dir string) (*VLog, error) {

	v := &VLog{
		files: make(map[uint32]*os.File),
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

		if fid > activeFid {
			activeFid = fid
		}

		file, err := os.Open(
			filepath.Join(dir, strconv.Itoa(int(fid))),
		)
		if err != nil {
			return nil, fmt.Errorf("open vlog: %w", err)
		}

		v.files[fid] = file
	}

	if activeFid == 0 {
		activeFid = 1
	}

	activePath := filepath.Join(dir, strconv.Itoa(int(activeFid)))

	file, err := os.OpenFile(
		activePath,
		os.O_CREATE|os.O_RDWR|os.O_APPEND,
		0644,
	)

	stat, err := file.Stat()
	if err != nil {
		file.Close()
		return nil, fmt.Errorf("open vlog: %w", err)
	}

	v.active = file
	v.activeFid = activeFid
	v.dir = dir
	v.offset = stat.Size()
	v.writer = bufio.NewWriterSize(file, 64*1024)

	return v, nil
}

func (v *VLog) Close() error {
	return v.active.Close()
}

func (v *VLog) Sync() error {
	if err := v.writer.Flush(); err != nil {
		return fmt.Errorf("vlog sync error: %w", err)
	}

	if err := v.active.Sync(); err != nil {
		return fmt.Errorf("vlog sync error: %w", err)
	}

	return nil
}

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

	var vp ValuePointer

	if err := EncodeEntry(entry, v.writer); err != nil {
		return vp, fmt.Errorf("append entry: %w", err)
	}

	totaLen := int64(8 + len(entry.key) + len(entry.value))

	vp.Offset = v.offset
	v.offset += totaLen

	vp.Fid = v.activeFid
	vp.Len = totaLen

	return vp, nil
}

func (v *VLog) Read(ptr ValuePointer) (Entry, error) {

	file, ok := v.files[ptr.Fid]
	if !ok {
		return Entry{}, fmt.Errorf("read vlog: file not found for fid %d", ptr.Fid)
	}

	r := io.NewSectionReader(file, ptr.Offset, ptr.Len)

	entry, err := DecodeEntry(r)
	if err != nil {
		return Entry{}, fmt.Errorf("read vlog: %w", err)
	}

	return entry, nil
}
