package vlog

import (
	"path/filepath"
	"testing"
)

func TestOpenVLog(t *testing.T) {
	path := filepath.Join(t.TempDir(), "vlog")

	v, err := OpenVLog(path)
	if err != nil {
		t.Fatalf("OpenVLog() error = %v", err)
	}
	defer v.Close()

	if v.offset != 0 {
		t.Fatalf("offset = %d, want 0", v.offset)
	}

	if v.fid != 1 {
		t.Fatalf("fid = %d, want 1", v.fid)
	}
}

func TestAppendRead(t *testing.T) {
	path := filepath.Join(t.TempDir(), "vlog")

	v, err := OpenVLog(path)
	if err != nil {
		t.Fatal(err)
	}
	defer v.Close()

	want := Entry{
		key:   []byte("hello"),
		value: []byte("world"),
	}

	ptr, err := v.Append(want)
	if err != nil {
		t.Fatalf("Append() error = %v", err)
	}

	if err := v.writer.Flush(); err != nil {
		t.Fatal(err)
	}

	got, err := v.Read(ptr)
	if err != nil {
		t.Fatalf("Read() error = %v", err)
	}

	if string(got.key) != string(want.key) {
		t.Errorf("key = %q, want %q", got.key, want.key)
	}

	if string(got.value) != string(want.value) {
		t.Errorf("value = %q, want %q", got.value, want.value)
	}
}

func TestAppendMultiple(t *testing.T) {
	path := filepath.Join(t.TempDir(), "vlog")

	v, err := OpenVLog(path)
	if err != nil {
		t.Fatal(err)
	}
	defer v.Close()

	entries := []Entry{
		{key: []byte("key1"), value: []byte("value1")},
		{key: []byte("key2"), value: []byte("value2")},
	}

	ptrs := make([]ValuePointer, len(entries))

	for i, entry := range entries {
		ptr, err := v.Append(entry)
		if err != nil {
			t.Fatal(err)
		}
		ptrs[i] = ptr
	}

	if err := v.writer.Flush(); err != nil {
		t.Fatal(err)
	}

	if ptrs[0].Offset != 0 {
		t.Errorf("first offset = %d, want 0", ptrs[0].Offset)
	}

	if ptrs[1].Offset != ptrs[0].Len {
		t.Errorf(
			"second offset = %d, want %d",
			ptrs[1].Offset,
			ptrs[0].Len,
		)
	}

	for i, ptr := range ptrs {
		got, err := v.Read(ptr)
		if err != nil {
			t.Fatal(err)
		}

		if string(got.key) != string(entries[i].key) {
			t.Errorf("entry %d key mismatch", i)
		}

		if string(got.value) != string(entries[i].value) {
			t.Errorf("entry %d value mismatch", i)
		}
	}
}

func TestReopenVLog(t *testing.T) {
	path := filepath.Join(t.TempDir(), "vlog")

	v, err := OpenVLog(path)
	if err != nil {
		t.Fatal(err)
	}

	entry := Entry{
		key:   []byte("hello"),
		value: []byte("world"),
	}

	ptr, err := v.Append(entry)
	if err != nil {
		t.Fatal(err)
	}

	if err := v.Sync(); err != nil {
		t.Fatal(err)
	}

	v.Close()

	v, err = OpenVLog(path)
	if err != nil {
		t.Fatal(err)
	}
	defer v.Close()

	if v.offset != ptr.Len {
		t.Errorf("offset = %d, want %d", v.offset, ptr.Len)
	}
}
