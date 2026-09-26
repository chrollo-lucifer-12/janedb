package memtable

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"path/filepath"
	"sync"
	"uuid"

	"github.com/janedb/internal/flags"
	"github.com/janedb/internal/keys"
	"github.com/janedb/internal/manifest"
	"github.com/janedb/internal/skl"

	"github.com/janedb/internal/sstable"
	"github.com/janedb/internal/vlog"
)

const MaxSize = 1 << 20

type Memtable struct {
	skl  *skl.Skiplist
	vlog *vlog.VLog

	numskeys int

	mu sync.RWMutex

	size uint64
}

func NewMemtable(v *vlog.VLog) (*Memtable, error) {

	skl := skl.NewSkiplist()

	values := v.Recover()

	for k, v := range values {
		skl.Insert(keys.InternalKey([]byte(k), v.Sequence, v.VType), v.Ptr)
	}

	return &Memtable{
		skl:  skl,
		vlog: v,
	}, nil
}

func (m *Memtable) Close() error {
	return m.vlog.Close()
}

func (m *Memtable) Put(key []byte, value []byte, sequence uint64, vType uint8) error {

	m.mu.Lock()
	defer m.mu.Unlock()

	tKey := keys.InternalKey(key, sequence, vType)

	ptr, err := m.vlog.Append(vlog.Entry{Key: key, Value: value, Sequence: sequence, VType: vType})

	if err != nil {
		return fmt.Errorf("memtable put: %w", err)
	}

	m.size += uint64(len(tKey))
	m.numskeys++
	m.skl.Insert(tKey, ptr)

	return nil
}

func (m *Memtable) UpdatePtr(key []byte, ptr vlog.ValuePointer) {

	m.mu.Lock()
	defer m.mu.Unlock()

	_, ok := m.skl.Search(key)

	if !ok {
		m.size += uint64(len(key))
	}

	m.skl.Insert(key, ptr)
}

func (m *Memtable) Get(key []byte, sequence uint64) (vlog.ValuePointer, bool) {

	m.mu.RLock()
	defer m.mu.RUnlock()

	tKey := keys.InternalKey(key, sequence, 1)

	ptr, internalKey := m.skl.Seek(tKey)
	if internalKey == nil {
		return vlog.ValuePointer{}, false
	}

	userKey := internalKey[:len(internalKey)-8]
	if !bytes.Equal(userKey, key) {
		return vlog.ValuePointer{}, false
	}

	tag := binary.LittleEndian.Uint64(internalKey[len(internalKey)-8:])
	vType := uint8(tag)

	if vType == 0 {
		return vlog.ValuePointer{}, false
	}

	return ptr, true
}

func (m *Memtable) Delete(key []byte, sequence uint64, vType uint8) error {

	m.mu.Lock()
	defer m.mu.Unlock()

	tKey := keys.InternalKey(key, sequence, vType)

	ptr, err := m.vlog.Append(vlog.Entry{Key: key, Value: nil, Sequence: sequence, VType: vType})
	if err != nil {
		return fmt.Errorf("memtable delete: %w", err)
	}

	m.skl.Insert(tKey, ptr)
	m.numskeys--

	return nil
}

func (m *Memtable) Flush() (manifest.SSTableMeta, error) {

	m.mu.Lock()
	defer m.mu.Unlock()

	it := skl.GetIterator(m.skl)

	sstmeta := manifest.SSTableMeta{
		Size:  0,
		Level: 0,
		ID:    uuid.New().String(),
	}

	sst, err := sstable.Create(filepath.Join(flags.SstDir, sstmeta.ID), m.numskeys)
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

		n, err := sst.WriteEntry(sstable.SSTableEntry{
			Key:   it.Key,
			Value: it.Value,
		})
		if err != nil {
			return manifest.SSTableMeta{}, err
		}

		sstmeta.Size += n
	}

	sst.WriteBF()

	sstmeta.Smallest = smallest
	sstmeta.Largest = largest

	if err := manifest.SaveManifest(sstmeta); err != nil {
		return manifest.SSTableMeta{}, err
	}

	m.skl = skl.NewSkiplist()
	m.numskeys = 0
	m.size = 0

	return sstmeta, nil
}

func (m *Memtable) IsOverflow() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.size+m.skl.GetSize() > MaxSize
}
