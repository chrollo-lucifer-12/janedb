package lsm

import (
	"bytes"
	"fmt"
	"os"

	"github.com/janedb/internal/flags"
	"github.com/janedb/internal/manifest"
	"github.com/janedb/internal/memtable"
	"github.com/janedb/internal/sstable"
	"github.com/janedb/internal/vlog"
)

const GCThreshold = 128

type LSM struct {
	m *memtable.Memtable
	v *vlog.VLog

	level []manifest.SSTableMeta

	gcStop chan struct{}
	gcDone chan struct{}
}

func OpenLSM() (*LSM, error) {

	if err := os.MkdirAll(flags.SstDir, 0755); err != nil {
		return nil, err
	}

	l := &LSM{}

	if err := l.recover(); err != nil {
		return nil, err
	}

	return l, nil
}

func (lsm *LSM) Close() error {
	if lsm.gcStop != nil {
		close(lsm.gcStop)
		<-lsm.gcDone
	}

	if err := saveTail(lsm.v.GetTail()); err != nil {
		return err
	}

	if err := saveHead(lsm.v.GetHead()); err != nil {
		return err
	}

	if err := lsm.v.Close(); err != nil {
		return err
	}

	return lsm.m.Close()
}

func (lsm *LSM) Put(key []byte, value []byte) error {

	if lsm.v.GetHead()-lsm.v.GetTail() >= GCThreshold {
		lsm.startGC()
	}

	if lsm.m.IsOverflow() {

		saveHead(lsm.v.GetHead())

		meta, err := lsm.m.Flush()
		if err != nil {
			return err
		}

		lsm.level = append(lsm.level, meta)
	}

	return lsm.m.Put(key, value)
}

func (lsm *LSM) Get(key []byte, buf []byte) bool {
	ptr, ok := lsm.m.Get([]byte(key))
	if ok {
		if err := lsm.v.ReadValue(ptr, buf); err != nil {
			return false
		}

		return true
	}

	ptr, err := lsm.getPtr(key)
	if err != nil {
		return false
	}

	lsm.v.ReadValue(ptr, buf)

	return true

}

func (lsm *LSM) Delete(key []byte) error {
	return lsm.m.Delete(key)
}

func (lsm *LSM) getPtr(key []byte) (vlog.ValuePointer, error) {

	for _, meta := range lsm.level {
		if bytes.Compare(key, meta.Smallest) >= 0 && bytes.Compare(key, meta.Largest) <= 0 {
			ptr, err := sstable.Read(key, meta.ID)
			if err != nil {
				return vlog.ValuePointer{}, err
			}

			return ptr, nil
		}
	}

	return vlog.ValuePointer{}, fmt.Errorf("key not found")
}
