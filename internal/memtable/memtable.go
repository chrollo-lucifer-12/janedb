package memtable

import (
	"fmt"

	"github.com/janedb/internal/skl"
	"github.com/janedb/internal/sstable"
	"github.com/janedb/internal/vlog"
)

const MaxSize = 1 << 20

type Memtable struct {
	skl  *skl.Skiplist
	vlog *vlog.VLog

	size uint64
}

func NewMemtable(dir string) (*Memtable, error) {

	v, err := vlog.OpenVLog(dir)
	if err != nil {
		return nil, err
	}

	skl := skl.NewSkiplist()

	values := v.Recover()

	fmt.Println(values)

	for k, v := range values {
		skl.Insert([]byte(k), v)
	}

	return &Memtable{
		skl:  skl,
		vlog: v,
	}, nil
}

func (m *Memtable) Put(key []byte, value []byte) error {

	if m.size+m.skl.GetSize() > MaxSize {
		if err := m.flush(); err != nil {
			return err
		}
	}

	ptr, err := m.vlog.Append(vlog.NewEntry(key, value))

	if err != nil {
		return fmt.Errorf("memtable put: %w", err)
	}

	m.size += uint64(len(key))
	m.skl.Insert(key, ptr)

	return nil
}

func (m *Memtable) Get(key []byte, buf []byte) bool {
	ptr, ok := m.skl.Search(key)

	if !ok {
		return false
	}

	if ptr.Len == 0 {
		return false
	}

	m.vlog.ReadValue(ptr, buf)

	return true
}

func (m *Memtable) Delete(key []byte) error {
	ptr, err := m.vlog.Append(vlog.NewEntry(key, nil))
	if err != nil {
		return fmt.Errorf("memtable delete: %w", err)
	}

	m.skl.Insert(key, ptr)

	return nil
}

func (m *Memtable) flush() error {
	it := skl.GetIterator(m.skl)

	sst, err := sstable.Create("")
	if err != nil {
		return err
	}

	for {
		if !it.GetNext() {
			break
		}

		if err := sst.Write(sstable.SSTableEntry{
			Key:   it.Key,
			Value: it.Value,
		}); err != nil {
			return err
		}
	}

	m.skl = skl.NewSkiplist()

	return nil
}
