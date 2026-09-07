package memtable

import (
	"fmt"

	"github.com/janedb/internal/skl"
	"github.com/janedb/internal/vlog"
)

type Memtable struct {
	skl  *skl.Skiplist
	vlog *vlog.VLog
}

func NewMemtable(dir string) (*Memtable, error) {

	v, err := vlog.OpenVLog(dir)
	if err != nil {
		return nil, err
	}

	skl := skl.NewSkiplist()

	values := v.Recover()

	for k, v := range values {
		skl.Insert([]byte(k), v)
	}

	return &Memtable{
		skl:  skl,
		vlog: v,
	}, nil
}

func (m *Memtable) Put(key []byte, value []byte) error {
	ptr, err := m.vlog.Append(vlog.NewEntry(key, value))

	if err != nil {
		return fmt.Errorf("memtable put: %w", err)
	}

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
