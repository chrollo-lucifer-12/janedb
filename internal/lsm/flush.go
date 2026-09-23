package lsm

import (
	"encoding/binary"
	"os"

	"github.com/janedb/internal/flags"
)

func saveHead(head int64) error {
	file, err := os.OpenFile(
		flags.VlogMarkers,
		os.O_CREATE|os.O_RDWR,
		0644,
	)
	if err != nil {
		return err
	}
	defer file.Close()

	if err := binary.Write(file, binary.BigEndian, head); err != nil {
		return err
	}

	return file.Sync()
}

func saveTail(tail int64) error {
	file, err := os.OpenFile(
		flags.VlogMarkers,
		os.O_CREATE|os.O_RDWR,
		0644,
	)
	if err != nil {
		return err
	}
	defer file.Close()

	if _, err := file.Seek(8, 0); err != nil {
		return err
	}

	if err := binary.Write(file, binary.BigEndian, tail); err != nil {
		return err
	}

	return file.Sync()
}
