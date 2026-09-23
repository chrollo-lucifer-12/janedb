package manifest

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"

	"github.com/janedb/internal/flags"
)

type SSTableMeta struct {
	ID       string
	Level    int
	Size     uint64
	Smallest []byte
	Largest  []byte
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
