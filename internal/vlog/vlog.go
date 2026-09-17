package vlog

import (
	"fmt"
	"os"
	"sync"

	"github.com/edsrzf/mmap-go"
	"golang.org/x/sys/unix"
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

func OpenVLog(path string, head, tail int64) (*VLog, error) {
	file, err := os.OpenFile(
		path,
		os.O_CREATE|os.O_RDWR,
		0644,
	)
	if err != nil {
		return nil, fmt.Errorf("open vlog: %w", err)
	}

	stat, err := file.Stat()
	if err != nil {
		file.Close()
		return nil, fmt.Errorf("stat vlog: %w", err)
	}

	if stat.Size() < MaxFileSize {
		if err := file.Truncate(MaxFileSize); err != nil {
			file.Close()
			return nil, fmt.Errorf("truncate vlog: %w", err)
		}
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
		file: file,
		data: data,
		head: head,
		tail: tail,
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

func (v *VLog) Sync() error {
	if err := v.data.Flush(); err != nil {
		return err
	}

	return v.file.Sync()
}

func (v *VLog) SetTail(tail int64) {
	v.tailmu.Lock()
	v.tail = tail
	v.tailmu.Unlock()
}

func (v *VLog) SetHead(head int64) {
	v.headmu.Lock()
	v.head = head
	v.headmu.Unlock()
}

func (v *VLog) GetHead() int64 {
	return v.head
}

func (v *VLog) GetTail() int64 {
	return v.tail
}

func (v *VLog) Reclaim(start, end int64) error {
	if end <= start {
		return nil
	}

	const (
		punchHole = 0x02
		keepSize  = 0x01
	)

	return unix.Fallocate(
		int(v.file.Fd()),
		punchHole|keepSize,
		start,
		end-start,
	)
}
