package lsm

import (
	"bytes"
	"os"

	"github.com/janedb/internal/flags"
	"github.com/janedb/internal/memtable"
	"github.com/janedb/internal/sstable"
)

type LSM struct {
	m     *memtable.Memtable
	level []sstable.SSTableMeta
}

func OpenLSM() (*LSM, error) {

	if err := os.MkdirAll(flags.SstDir, 0755); err != nil {
		return nil, err
	}

	var err error
	l := &LSM{}

	l.m, err = memtable.NewMemtable(flags.VlogDir)
	if err != nil {
		return nil, err
	}

	l.level, err = sstable.RecoverManifest()

	return l, nil
}

func (lsm *LSM) Close() error {
	return lsm.Close()
}

func (lsm *LSM) Put(key []byte, value []byte) error {

	if lsm.m.IsOverflow() {
		meta, err := lsm.m.Flush()
		if err != nil {
			return err
		}

		lsm.level = append(lsm.level, meta)
	}

	return lsm.m.Put(key, value)
}

func (lsm *LSM) Get(key []byte, buf []byte) bool {
	ok := lsm.m.Get([]byte(key), buf)
	if ok {
		return true
	}

	for _, meta := range lsm.level {
		if bytes.Compare(key, meta.Smallest) >= 0 && bytes.Compare(key, meta.Largest) <= 0 {
			ptr, err := sstable.Read(key, meta.ID)
			if err != nil {
				return false
			}

			lsm.m.ReadValue(ptr, buf)

			return true
		}
	}

	return false
}

func (lsm *LSM) Delete(key []byte) error {
	return lsm.m.Delete(key)
}
