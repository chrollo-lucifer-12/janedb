package sstable

import (
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"

	"github.com/janedb/internal/flags"
	"github.com/janedb/internal/vlog"
)

const VPSize = 16

type SSTableEntry struct {
	Key   []byte
	Value vlog.ValuePointer
}

type SSTable struct {
	file *os.File
}

func Create(path string) (*SSTable, error) {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return nil, fmt.Errorf("create sstable: %w", err)
	}

	return &SSTable{
		file: file,
	}, nil
}

func Read(key []byte, id string) (vlog.ValuePointer, error) {
	path := filepath.Join(flags.SstDir, id)

	start := int64(0)

	file, err := os.Open(path)
	if err != nil {
		return vlog.ValuePointer{}, fmt.Errorf("sstable read: %w", err)
	}

	stat, err := file.Stat()
	if err != nil {
		return vlog.ValuePointer{}, fmt.Errorf("sstable read: %w", err)
	}

	for start < stat.Size() {
		var keyLen uint32

		err := binary.Read(file, binary.BigEndian, &keyLen)
		if err != nil {
			break
		}

		start += 4 + int64(keyLen)

	}

	return vlog.ValuePointer{}, nil
}

func (s *SSTable) Write(entry SSTableEntry) (uint64, error) {

	keyLen := uint32(len(entry.Key))

	if err := binary.Write(s.file, binary.BigEndian, keyLen); err != nil {
		return 0, err
	}

	s.file.Write(entry.Key)

	binary.Write(s.file, binary.BigEndian, entry.Value.Offset)
	binary.Write(s.file, binary.BigEndian, entry.Value.Len)

	return 20 + uint64(keyLen), nil
}
