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

	if v.activeFid != 1 {
		t.Fatalf("activeFid = %d, want 1", v.activeFid)
	}

	if v.offset != 0 {
		t.Fatalf("offset = %d, want 0", v.offset)
	}

	if v.active == nil {
		t.Fatal("active file is nil")
	}

	if len(v.files) != 1 {
		t.Fatalf("files = %d, want 1", len(v.files))
	}

	if v.files[1] == nil {
		t.Fatal("files[1] is nil")
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

	if err := v.Sync(); err != nil {
		t.Fatal(err)
	}

	if ptr.Fid != 1 {
		t.Errorf("ptr.Fid = %d, want 1", ptr.Fid)
	}

	if ptr.Offset != 0 {
		t.Errorf("ptr.Offset = %d, want 0", ptr.Offset)
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

	if err := v.Sync(); err != nil {
		t.Fatal(err)
	}

	if ptrs[0].Fid != 1 || ptrs[1].Fid != 1 {
		t.Errorf("expected both entries in segment 1")
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

func TestRotate(t *testing.T) {
	path := filepath.Join(t.TempDir(), "vlog")

	v, err := OpenVLog(path)
	if err != nil {
		t.Fatal(err)
	}
	defer v.Close()

	if err := v.Rotate(); err != nil {
		t.Fatalf("Rotate() error = %v", err)
	}

	if v.activeFid != 2 {
		t.Fatalf("activeFid = %d, want 2", v.activeFid)
	}

	if v.offset != 0 {
		t.Fatalf("offset = %d, want 0", v.offset)
	}

	if v.active == nil {
		t.Fatal("active file is nil")
	}

	if len(v.files) != 2 {
		t.Fatalf("files = %d, want 2", len(v.files))
	}

	if v.files[2] == nil {
		t.Fatal("files[2] is nil")
	}
}

func TestReadAcrossSegments(t *testing.T) {
	path := filepath.Join(t.TempDir(), "vlog")

	v, err := OpenVLog(path)
	if err != nil {
		t.Fatal(err)
	}
	defer v.Close()

	first := Entry{
		key:   []byte("first"),
		value: []byte("value1"),
	}

	firstPtr, err := v.Append(first)
	if err != nil {
		t.Fatal(err)
	}

	if err := v.Rotate(); err != nil {
		t.Fatal(err)
	}

	second := Entry{
		key:   []byte("second"),
		value: []byte("value2"),
	}

	secondPtr, err := v.Append(second)
	if err != nil {
		t.Fatal(err)
	}

	if err := v.Sync(); err != nil {
		t.Fatal(err)
	}

	if firstPtr.Fid != 1 {
		t.Errorf("firstPtr.Fid = %d, want 1", firstPtr.Fid)
	}

	if secondPtr.Fid != 2 {
		t.Errorf("secondPtr.Fid = %d, want 2", secondPtr.Fid)
	}

	gotFirst, err := v.Read(firstPtr)
	if err != nil {
		t.Fatalf("Read(firstPtr) error = %v", err)
	}

	gotSecond, err := v.Read(secondPtr)
	if err != nil {
		t.Fatalf("Read(secondPtr) error = %v", err)
	}

	if string(gotFirst.value) != string(first.value) {
		t.Errorf("first value = %q, want %q", gotFirst.value, first.value)
	}

	if string(gotSecond.value) != string(second.value) {
		t.Errorf("second value = %q, want %q", gotSecond.value, second.value)
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

	if err := v.Close(); err != nil {
		t.Fatal(err)
	}

	v, err = OpenVLog(path)
	if err != nil {
		t.Fatal(err)
	}
	defer v.Close()

	if v.activeFid != 1 {
		t.Errorf("activeFid = %d, want 1", v.activeFid)
	}

	if v.offset != ptr.Len {
		t.Errorf("offset = %d, want %d", v.offset, ptr.Len)
	}

	got, err := v.Read(ptr)
	if err != nil {
		t.Fatalf("Read() after reopen error = %v", err)
	}

	if string(got.key) != string(entry.key) {
		t.Errorf("key = %q, want %q", got.key, entry.key)
	}

	if string(got.value) != string(entry.value) {
		t.Errorf("value = %q, want %q", got.value, entry.value)
	}
}
