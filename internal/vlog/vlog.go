package vlog

import (
	"fmt"
	"os"
	"sync"

	"github.com/edsrzf/mmap-go"
)

const MaxFileSize = 1 << 20

type logFile struct {
	file *os.File
	data mmap.MMap
}

type VLog struct {
	logFile

	headmu sync.Mutex
	head   int64

	tailmu sync.Mutex
	tail   int64
}

type ValuePointer struct {
	Offset int64
	Len    int64
}

func OpenVLog(path string) (*VLog, error) {
	file, err := os.OpenFile(
		path,
		os.O_CREATE|os.O_RDWR,
		0644,
	)
	if err != nil {
		return nil, fmt.Errorf("open vlog: %w", err)
	}

	if err := file.Truncate(MaxFileSize); err != nil {
		file.Close()
		return nil, fmt.Errorf("truncate vlog: %w", err)
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
		return nil, fmt.Errorf("mmap vlog: %w", err)
	}

	v := &VLog{
		logFile: logFile{
			file: file,
			data: data,
		},
		head: 0,
		tail: 0,
	}

	return v, nil
}

func (v *VLog) Close() error {
	if v.data != nil {
		if err := v.data.Flush(); err != nil {
			return fmt.Errorf("flush vlog: %w", err)
		}

		if err := v.data.Unmap(); err != nil {
			return fmt.Errorf("unmap vlog: %w", err)
		}
	}

	if v.file != nil {
		if err := v.file.Close(); err != nil {
			return fmt.Errorf("close vlog: %w", err)
		}
	}

	return nil
}
