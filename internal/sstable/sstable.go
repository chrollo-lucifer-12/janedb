package sstable

import (
	"encoding/binary"
	"fmt"
	"os"

	"github.com/janedb/internal/vlog"
)

const VPSize = 20

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

func (s *SSTable) Read(key []byte) (vlog.ValuePointer, error) {
	dir := ""

	entries, err := os.ReadDir(dir)
	if err != nil {
		return vlog.ValuePointer{}, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		start := int64(0)

		file, err := os.Open(entry.Name())
		if err != nil {
			continue
		}

		stat, err := file.Stat()
		if err != nil {
			continue
		}

		for start < stat.Size() {
			var keyLen uint32

			err := binary.Read(file, binary.BigEndian, &keyLen)
			if err != nil {
				break
			}

			start += 4 + int64(keyLen)

		}
	}

	return vlog.ValuePointer{}, nil
}

func (s *SSTable) Write(entry SSTableEntry) error {

	keyLen := uint32(len(entry.Key))

	if err := binary.Write(s.file, binary.BigEndian, keyLen); err != nil {
		return err
	}

	s.file.Write(entry.Key)

	binary.Write(s.file, binary.BigEndian, entry.Value.Fid)
	binary.Write(s.file, binary.BigEndian, entry.Value.Offset)
	binary.Write(s.file, binary.BigEndian, entry.Value.Len)

	return nil
}
