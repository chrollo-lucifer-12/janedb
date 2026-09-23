package memtable

import (
	"fmt"
	"path/filepath"
	"uuid"

	"github.com/janedb/internal/flags"
	"github.com/janedb/internal/manifest"
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

func NewMemtable(v *vlog.VLog) (*Memtable, error) {

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

func (m *Memtable) Close() error {
	return m.vlog.Close()
}

func (m *Memtable) Put(key []byte, value []byte) error {

	ptr, err := m.vlog.Append(vlog.Entry{Key: key, Value: value})

	if err != nil {
		return fmt.Errorf("memtable put: %w", err)
	}

	m.size += uint64(len(key))
	m.skl.Insert(key, ptr)

	return nil
}

func (m *Memtable) UpdatePtr(key []byte, ptr vlog.ValuePointer) {
	_, ok := m.skl.Search(key)

	if !ok {
		m.size += uint64(len(key))
	}

	m.skl.Insert(key, ptr)
}

func (m *Memtable) Get(key []byte) (vlog.ValuePointer, bool) {
	ptr, ok := m.skl.Search(key)

	if ptr.Len == 0 {
		return ptr, false
	}

	return ptr, ok
}

func (m *Memtable) Delete(key []byte) error {
	ptr, err := m.vlog.Append(vlog.Entry{Key: key, Value: nil})
	if err != nil {
		return fmt.Errorf("memtable delete: %w", err)
	}

	m.skl.Insert(key, ptr)

	return nil
}

func (m *Memtable) Flush() (manifest.SSTableMeta, error) {
	it := skl.GetIterator(m.skl)

	sstmeta := manifest.SSTableMeta{
		Size:  0,
		Level: 0,
		ID:    uuid.New().String(),
	}

	sst, err := sstable.Create(filepath.Join(flags.SstDir, sstmeta.ID))
	if err != nil {
		return manifest.SSTableMeta{}, err
	}

	var smallest, largest []byte

	first := true

	for {
		if !it.GetNext() {
			break
		}

		if first {
			smallest = append([]byte(nil), it.Key...)
			first = false
		}
		largest = append(largest[:0], it.Key...)

		n, err := sst.Write(sstable.SSTableEntry{
			Key:   it.Key,
			Value: it.Value,
		})
		if err != nil {
			return manifest.SSTableMeta{}, err
		}

		sstmeta.Size += n
	}

	sstmeta.Smallest = smallest
	sstmeta.Largest = largest

	if err := manifest.SaveManifest(sstmeta); err != nil {
		return manifest.SSTableMeta{}, err
	}

	m.skl = skl.NewSkiplist()

	return manifest.SSTableMeta{}, nil
}

func (m *Memtable) IsOverflow() bool {
	return m.size+m.skl.GetSize() > MaxSize
}
