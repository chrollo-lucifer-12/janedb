package sstable

import (
	"encoding/binary"
	"fmt"
	"os"

	"github.com/janedb/internal/vlog"
)

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

func (s *SSTable) Read() {

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
