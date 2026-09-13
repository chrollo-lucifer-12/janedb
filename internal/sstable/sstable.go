package sstable

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/janedb/internal/flags"
	"github.com/janedb/internal/vlog"
)

const VPSize = 20

type SSTableMeta struct {
	ID       string
	Level    int
	Size     uint64
	Smallest []byte
	Largest  []byte
}

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

	binary.Write(s.file, binary.BigEndian, entry.Value.Fid)
	binary.Write(s.file, binary.BigEndian, entry.Value.Offset)
	binary.Write(s.file, binary.BigEndian, entry.Value.Len)

	return 24 + uint64(keyLen), nil
}

func SaveManifest(meta SSTableMeta) error {
	f, err := os.OpenFile(
		flags.SstManifest,
		os.O_CREATE|os.O_WRONLY|os.O_APPEND,
		0644,
	)
	if err != nil {
		return fmt.Errorf("save manifest: %w", err)
	}
	defer f.Close()

	if err := binary.Write(f, binary.BigEndian, uint32(len(meta.ID))); err != nil {
		return fmt.Errorf("save manifest: %w", err)
	}

	if _, err := io.WriteString(f, meta.ID); err != nil {
		return fmt.Errorf("save manifest: %w", err)
	}

	if err := binary.Write(f, binary.BigEndian, int32(meta.Level)); err != nil {
		return fmt.Errorf("save manifest: %w", err)
	}

	if err := binary.Write(f, binary.BigEndian, meta.Size); err != nil {
		return fmt.Errorf("save manifest: %w", err)
	}

	if err := binary.Write(f, binary.BigEndian, uint32(len(meta.Smallest))); err != nil {
		return fmt.Errorf("save manifest: %w", err)
	}

	if _, err := f.Write(meta.Smallest); err != nil {
		return fmt.Errorf("save manifest: %w", err)
	}

	if err := binary.Write(f, binary.BigEndian, uint32(len(meta.Largest))); err != nil {
		return fmt.Errorf("save manifest: %w", err)
	}

	if _, err := f.Write(meta.Largest); err != nil {
		return fmt.Errorf("save manifest: %w", err)
	}

	return nil
}

func RecoverManifest() ([]SSTableMeta, error) {
	f, err := os.Open(
		flags.SstManifest,
	)
	if err != nil {
		return nil, fmt.Errorf("save manifest: %w", err)
	}
	defer f.Close()

	var metas []SSTableMeta

	for {
		var idLen uint32

		if err := binary.Read(f, binary.BigEndian, &idLen); err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("recover manifest: read id length: %w", err)
		}

		id := make([]byte, idLen)
		if _, err := io.ReadFull(f, id); err != nil {
			return nil, fmt.Errorf("recover manifest: read id: %w", err)
		}

		var level int32
		if err := binary.Read(f, binary.BigEndian, &level); err != nil {
			return nil, fmt.Errorf("recover manifest: read level: %w", err)
		}

		var size uint64
		if err := binary.Read(f, binary.BigEndian, &size); err != nil {
			return nil, fmt.Errorf("recover manifest: read size: %w", err)
		}

		var smallestLen uint32
		if err := binary.Read(f, binary.BigEndian, &smallestLen); err != nil {
			return nil, fmt.Errorf("recover manifest: read smallest length: %w", err)
		}

		smallest := make([]byte, smallestLen)
		if _, err := io.ReadFull(f, smallest); err != nil {
			return nil, fmt.Errorf("recover manifest: read smallest: %w", err)
		}

		var largestLen uint32
		if err := binary.Read(f, binary.BigEndian, &largestLen); err != nil {
			return nil, fmt.Errorf("recover manifest: read largest length: %w", err)
		}

		largest := make([]byte, largestLen)
		if _, err := io.ReadFull(f, largest); err != nil {
			return nil, fmt.Errorf("recover manifest: read largest: %w", err)
		}

		metas = append(metas, SSTableMeta{
			ID:       string(id),
			Level:    int(level),
			Size:     size,
			Smallest: smallest,
			Largest:  largest,
		})
	}

	return metas, nil

}
