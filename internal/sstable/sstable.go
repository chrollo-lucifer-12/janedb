package sstable

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/janedb/internal/bloom"
	"github.com/janedb/internal/flags"
	"github.com/janedb/internal/vlog"
)

const VPSize = 16
const FooterSize = 8

type SSTableEntry struct {
	Key   []byte
	Value vlog.ValuePointer
}

type SSTable struct {
	file *os.File
	bf   *bloom.BF
}

func Create(path string, numkeys int) (*SSTable, error) {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return nil, fmt.Errorf("create sstable: %w", err)
	}

	return &SSTable{
		file: file,
		bf:   bloom.NewBloomFilter(numkeys),
	}, nil
}

func (s *SSTable) WriteBF() error {
	offset, err := s.file.Seek(0, io.SeekCurrent)
	if err != nil {
		return err
	}

	if err := s.bf.Marshal(s.file); err != nil {
		return err
	}

	return s.writeFooter(uint64(offset))
}

func (s *SSTable) writeFooter(bloomOffset uint64) error {
	var buf [8]byte

	binary.LittleEndian.PutUint64(buf[:], bloomOffset)

	_, err := s.file.Write(buf[:])
	return err
}

func readBloomOffset(file *os.File) (uint64, error) {
	if _, err := file.Seek(-FooterSize, io.SeekEnd); err != nil {
		return 0, err
	}

	var buf [8]byte

	if _, err := io.ReadFull(file, buf[:]); err != nil {
		return 0, err
	}

	return binary.LittleEndian.Uint64(buf[:]), nil
}

func Read(userkey []byte, internalkey []byte, id string) (vlog.ValuePointer, error) {
	path := filepath.Join(flags.SstDir, id)

	file, err := os.Open(path)
	if err != nil {
		return vlog.ValuePointer{}, fmt.Errorf("sstable read: %w", err)
	}
	defer file.Close()

	bloomOffset, err := readBloomOffset(file)
	if err != nil {
		return vlog.ValuePointer{}, err
	}

	if _, err := file.Seek(int64(bloomOffset), io.SeekStart); err != nil {
		return vlog.ValuePointer{}, err
	}

	bf := &bloom.BF{}
	if err := bf.Unmarshal(file); err != nil {
		return vlog.ValuePointer{}, err
	}

	if !bf.MayContain(userkey) {
		return vlog.ValuePointer{}, fmt.Errorf("key not found")
	}

	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return vlog.ValuePointer{}, err
	}

	var start int64

	for start < int64(bloomOffset) {
		var keyLen uint32

		if err := binary.Read(file, binary.BigEndian, &keyLen); err != nil {
			return vlog.ValuePointer{}, err
		}

		start += 4

		if start+int64(keyLen)+VPSize > int64(bloomOffset) {
			return vlog.ValuePointer{}, fmt.Errorf("corrupt sstable")
		}

		keyBuf := make([]byte, keyLen)

		if _, err := io.ReadFull(file, keyBuf); err != nil {
			return vlog.ValuePointer{}, err
		}

		start += int64(keyLen)

		var offset uint64
		var length uint64

		if err := binary.Read(file, binary.BigEndian, &offset); err != nil {
			return vlog.ValuePointer{}, err
		}

		if err := binary.Read(file, binary.BigEndian, &length); err != nil {
			return vlog.ValuePointer{}, err
		}

		start += VPSize

		if bytes.Equal(keyBuf, internalkey) {
			return vlog.ValuePointer{
				Offset: int64(offset),
				Len:    int64(length),
			}, nil
		}
	}

	return vlog.ValuePointer{}, fmt.Errorf("key not found")
}

func (s *SSTable) WriteEntry(entry SSTableEntry) (uint64, error) {

	keyLen := uint32(len(entry.Key))

	userKey := entry.Key[:len(entry.Key)-8]
	s.bf.Add(userKey)

	if err := binary.Write(s.file, binary.BigEndian, keyLen); err != nil {
		return 0, err
	}

	if _, err := s.file.Write(entry.Key); err != nil {
		return 0, err
	}

	if err := binary.Write(s.file, binary.BigEndian, entry.Value.Offset); err != nil {
		return 0, err
	}

	if err := binary.Write(s.file, binary.BigEndian, entry.Value.Len); err != nil {
		return 0, err
	}

	return 20 + uint64(keyLen), nil
}
