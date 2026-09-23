package lsm

import (
	"encoding/binary"
	"errors"
	"os"

	"github.com/janedb/internal/flags"
	"github.com/janedb/internal/manifest"
	"github.com/janedb/internal/memtable"
	"github.com/janedb/internal/vlog"
)

func (l *LSM) recover() error {
	var err error

	tail, head, err := getMarkers()

	l.v, err = vlog.OpenVLog(flags.VlogDir, head, tail)
	if err != nil {
		return err
	}

	l.m, err = memtable.NewMemtable(l.v)
	if err != nil {
		return err
	}

	l.level, err = manifest.RecoverManifest()

	return nil
}

func getMarkers() (int64, int64, error) {
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
