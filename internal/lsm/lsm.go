package lsm

import (
	"bytes"
	"fmt"
	"os"

	"github.com/janedb/internal/flags"
	"github.com/janedb/internal/memtable"
	"github.com/janedb/internal/sstable"
	"github.com/janedb/internal/vlog"
)

type LSM struct {
	m *memtable.Memtable
	v *vlog.VLog

	level []sstable.SSTableMeta
}

func OpenLSM() (*LSM, error) {

	if err := os.MkdirAll(flags.SstDir, 0755); err != nil {
		return nil, err
	}

	var err error
	l := &LSM{}

	tail, head, err := GetMarkers()

	l.v, err = vlog.OpenVLog(flags.VlogDir, head, tail)
	if err != nil {
		return nil, err
	}

	l.m, err = memtable.NewMemtable(l.v)
	if err != nil {
		return nil, err
	}

	l.level, err = sstable.RecoverManifest()

	return l, nil
}

func (lsm *LSM) Close() error {
	return lsm.m.Close()
}

func (lsm *LSM) Put(key []byte, value []byte) error {

	if lsm.m.IsOverflow() {

		SaveHead(lsm.v.GetHead())

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
	if !ok {
		return false
	}

	if err := lsm.v.ReadValue(ptr, buf); err != nil {
		return false
	}

	ptr, err := lsm.getPtr(key)
	if err != nil {
		return false
	}

	lsm.m.ReadValue(ptr, buf)

	return true

}

func (lsm *LSM) Delete(key []byte) error {
	return lsm.m.Delete(key)
}

func (lsm *LSM) getPtr(key []byte) (vlog.ValuePointer, error) {

	ptr, ok := lsm.m.Get([]byte(key))
	if ok {
		return ptr, nil
	}

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
