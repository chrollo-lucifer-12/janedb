package db

import "github.com/janedb/internal/memtable"

type DB struct {
	m *memtable.Memtable
}

func NewDB() (*DB, error) {

	dir := "data/vlog"

	m, err := memtable.NewMemtable(dir)
	if err != nil {
		return nil, err
	}

	return &DB{
		m: m,
	}, nil
}

func (d *DB) Put(key []byte, value []byte) error {
	return d.m.Put(key, value)
}

func (d *DB) Get(key []byte, buf []byte) bool {
	return d.m.Get([]byte(key), buf)
}
