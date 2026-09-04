package vlog

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type VLog struct {
	file   *os.File
	writer *bufio.Writer
	fid    uint32
	offset int64
}

type ValuePointer struct {
	Fid    uint32
	Offset int64
	Len    int64
}

func OpenVLog(path string) (*VLog, error) {

	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("open vlog: %v", err)
	}

	file, err := os.OpenFile(absPath, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return nil, fmt.Errorf("open vlog: %v", err)
	}

	stat, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("open vlog: %v", err)
	}

	return &VLog{
		file:   file,
		writer: bufio.NewWriterSize(file, 64*1024),
		offset: stat.Size(),
		fid:    1,
	}, nil

}

func (v *VLog) Close() error {
	return v.file.Close()
}

func (v *VLog) Sync() error {
	if err := v.writer.Flush(); err != nil {
		return fmt.Errorf("vlog sync error: %w", err)
	}

	if err := v.file.Sync(); err != nil {
		return fmt.Errorf("vlog sync error: %w", err)
	}

	return nil
}

func (v *VLog) Append(entry Entry) (ValuePointer, error) {

	var vp ValuePointer

	if err := EncodeEntry(entry, v.writer); err != nil {
		return vp, fmt.Errorf("append entry: %w", err)
	}

	totaLen := int64(8 + len(entry.key) + len(entry.value))

	v.offset += totaLen

	vp.Offset = v.offset
	vp.Fid = v.fid
	vp.Len = totaLen

	return vp, nil
}

func (v *VLog) Read(ptr ValuePointer) (Entry, error) {
	r := io.NewSectionReader(v.file, ptr.Offset, ptr.Len)

	entry, err := DecodeEntry(r)
	if err != nil {
		return Entry{}, fmt.Errorf("read vlog: %w", err)
	}

	return entry, nil
}
