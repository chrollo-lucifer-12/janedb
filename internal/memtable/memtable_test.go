package memtable

import (
	"bytes"
	"fmt"
	"testing"
)

func TestGet(t *testing.T) {

	dir := t.TempDir()

	m, err := NewMemtable(dir)

	if err != nil {
		t.Fatalf(err.Error())
	}

	for i := 0; i < 100; i++ {
		key := []byte(fmt.Sprintf("key+%d", i))
		val := []byte(fmt.Sprintf("value+%d", i))

		m.Put(key, val)

		buf := make([]byte, len(val))

		ok := m.Get(key, buf)
		if !ok {
			t.Fatalf("get failed for %q: %v", key, err)
		}

		if !bytes.Equal(val, buf) {
			t.Fatalf("key %q: want %q, got %q", key, val, buf)
		}
	}
}

func TestDelete(t *testing.T) {

	dir := t.TempDir()

	m, err := NewMemtable(dir)
	if err != nil {
		t.Fatal(err)
	}

	key := []byte("foo")
	value := []byte("bar")

	if err := m.Put(key, value); err != nil {
		t.Fatal(err)
	}

	buf := make([]byte, len(value))
	if !m.Get(key, buf) {
		t.Fatal("expected key to exist")
	}

	if err := m.Delete(key); err != nil {
		t.Fatal(err)
	}

	if m.Get(key, buf) {
		t.Fatal("expected key to be deleted")
	}
}
