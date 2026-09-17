package lsm

import (
	"encoding/binary"
	"errors"
	"os"

	"github.com/janedb/internal/flags"
)

func SaveHead(head int64) error {
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

func SaveTail(tail int64) error {
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

func GetMarkers() (int64, int64, error) {
	file, err := os.Open(flags.VlogMarkers)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return 0, 0, nil
		}
		return -1, -1, err
	}

	var (
		tail int64
		head int64
	)

	if err := binary.Read(file, binary.BigEndian, &tail); err != nil {
		return -1, -1, err
	}

	if err := binary.Read(file, binary.BigEndian, &head); err != nil {
		return -1, -1, err
	}

	return tail, head, nil
}

func (lsm *LSM) VLogHead() int64 { return lsm.v.GetHead() }

func (lsm *LSM) VLogTail() int64 { return lsm.v.GetTail() }
