package vlog

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"sync"
)

const MaxFileSize = 1 << 20

type VLog struct {
	dir string

	mu     sync.RWMutex
	active *os.File
	writer *bufio.Writer

	activeFid uint32
	offset    int64

	files map[uint32]*os.File

	readQueue chan readJob
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

type readJob struct {
	index   int
	ptr     ValuePointer
	results []ReadResult
	wg      *sync.WaitGroup
}

func OpenVLog(dir string) (*VLog, error) {

	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("open vlog: %w", err)
	}

	v := &VLog{
		dir:   dir,
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

		file, err := os.Open(
			filepath.Join(dir, strconv.Itoa(int(fid))),
		)
		if err != nil {
			return nil, fmt.Errorf("open vlog: %w", err)
		}

		v.files[fid] = file

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

		v.files[1] = file
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
	v.offset = stat.Size()
	v.writer = bufio.NewWriterSize(file, 64*1024)
	v.readQueue = make(chan readJob, 1024)

	for i := 0; i < 8; i++ {
		go v.readWorker()
	}

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
