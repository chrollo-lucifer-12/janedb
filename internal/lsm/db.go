package lsm

import (
	"fmt"
	"os"

	"github.com/janedb/internal/flags"
	"github.com/janedb/internal/keys"
	"github.com/janedb/internal/manifest"
	"github.com/janedb/internal/memtable"

	"github.com/janedb/internal/sstable"
	"github.com/janedb/internal/vlog"
)

type ValueType uint8

const GCThreshold = 128

const (
	TypeDeletion ValueType = 0
	TypeAddition ValueType = 1
)

type LSM struct {
	m *memtable.Memtable
	v *vlog.VLog

	sequence uint64

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

	if err := saveSequence(int64(lsm.sequence)); err != nil {
		return err
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

	lsm.sequence++
	return lsm.m.Put(key, value, lsm.sequence, uint8(TypeAddition))
}

func (lsm *LSM) Get(key []byte, buf []byte) (int, bool) {
	ptr, ok := lsm.m.Get([]byte(key), lsm.sequence)
	if ok {
		if err := lsm.v.ReadValue(ptr, buf); err != nil {
			return 0, false
		}

		return int(ptr.Len), true
	}

	ptr, err := lsm.getPtr(key)
	if err != nil {
		return 0, false
	}

	lsm.v.ReadValue(ptr, buf)

	return int(ptr.Len), true
}

func (lsm *LSM) Delete(key []byte) error {
	lsm.sequence++
	return lsm.m.Delete(key, lsm.sequence, uint8(TypeAddition))
}

func (lsm *LSM) getPtr(key []byte) (vlog.ValuePointer, error) {

	lookup := keys.InternalKey(key, lsm.sequence, 1)

	for _, meta := range lsm.level {
		if keys.CompareInternalKey(lookup, meta.Smallest) >= 0 && keys.CompareInternalKey(lookup, meta.Largest) <= 0 {
			ptr, err := sstable.Read(key, lookup, meta.ID)
			if err != nil {
				return vlog.ValuePointer{}, err
			}
			return ptr, nil
		}
	}

	return vlog.ValuePointer{}, fmt.Errorf("key not found")
}
