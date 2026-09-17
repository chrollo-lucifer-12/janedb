package vlog

// func BenchmarkVLogAppend(b *testing.B) {
// 	dir := b.TempDir()

// 	v, err := OpenVLog(dir)
// 	if err != nil {
// 		b.Fatal(err)
// 	}
// 	defer v.Close()

// 	entry := Entry{
// 		key:   []byte("benchmark-key"),
// 		value: []byte("benchmark-value"),
// 	}

// 	b.ResetTimer()

// 	for i := 0; i < b.N; i++ {
// 		if _, err := v.Append(entry); err != nil {
// 			b.Fatal(err)
// 		}
// 	}
// }

// func BenchmarkVLogRead(b *testing.B) {
// 	dir := b.TempDir()

// 	v, err := OpenVLog(dir)
// 	if err != nil {
// 		b.Fatal(err)
// 	}
// 	defer v.Close()

// 	const numPtrs = 1000

// 	ptrs := make([]ValuePointer, numPtrs)
// 	bufs := make([][]byte, numPtrs)

// 	for i := 0; i < numPtrs; i++ {
// 		entry := Entry{
// 			key:   []byte(fmt.Sprintf("key-%d", i)),
// 			value: []byte(fmt.Sprintf("value-%d", i)),
// 		}

// 		ptr, err := v.Append(entry)
// 		if err != nil {
// 			b.Fatal(err)
// 		}

// 		ptrs[i] = ptr
// 		bufs[i] = make([]byte, ptr.Len)
// 	}

// 	if err := v.Sync(); err != nil {
// 		b.Fatal(err)
// 	}

// 	b.ResetTimer()

// 	for i := 0; i < b.N; i++ {
// 		for j := 0; j < numPtrs; j++ {
// 			if err := v.ReadValue(ptrs[j], bufs[j]); err != nil {
// 				b.Fatal(err)
// 			}
// 		}
// 	}
// }

// func TestOpenVLog(t *testing.T) {
// 	path := filepath.Join(t.TempDir(), "vlog")

// 	v, err := OpenVLog(path)
// 	if err != nil {
// 		t.Fatalf("OpenVLog() error = %v", err)
// 	}
// 	defer v.Close()

// 	if v.activeFid != 1 {
// 		t.Fatalf("activeFid = %d, want 1", v.activeFid)
// 	}

// 	if v.offset != 0 {
// 		t.Fatalf("offset = %d, want 0", v.offset)
// 	}

// 	if len(v.files) != 1 {
// 		t.Fatalf("files = %d, want 1", len(v.files))
// 	}

// 	if v.files[1] == nil {
// 		t.Fatal("files[1] is nil")
// 	}
// }

// func TestAppendRead(t *testing.T) {
// 	path := filepath.Join(t.TempDir(), "vlog")

// 	v, err := OpenVLog(path)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	defer v.Close()

// 	want := Entry{
// 		key:   []byte("hello"),
// 		value: []byte("world"),
// 	}

// 	ptr, err := v.Append(want)
// 	if err != nil {
// 		t.Fatalf("Append() error = %v", err)
// 	}

// 	if err := v.Sync(); err != nil {
// 		t.Fatal(err)
// 	}

// 	if ptr.Fid != 1 {
// 		t.Errorf("ptr.Fid = %d, want 1", ptr.Fid)
// 	}

// 	expectedOffset := int64(8 + len(want.key))

// 	if ptr.Offset != expectedOffset {
// 		t.Errorf("ptr.Offset = %d, want 0", ptr.Offset)
// 	}

// 	result := make([]byte, len(want.value))
// 	err = v.ReadValue(ptr, result)

// 	if err != nil {
// 		t.Errorf(err.Error())
// 	}

// 	if string(result) != string(want.value) {
// 		t.Errorf("value = %q, want %q", result, want.value)
// 	}

// }

// func TestAppendMultiple(t *testing.T) {
// 	path := filepath.Join(t.TempDir(), "vlog")

// 	v, err := OpenVLog(path)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	defer v.Close()

// 	entries := []Entry{
// 		{key: []byte("key1"), value: []byte("value1")},
// 		{key: []byte("key2"), value: []byte("value2")},
// 	}

// 	ptrs := make([]ValuePointer, len(entries))

// 	for i, entry := range entries {
// 		ptr, err := v.Append(entry)
// 		if err != nil {
// 			t.Fatal(err)
// 		}
// 		ptrs[i] = ptr
// 	}

// 	if err := v.Sync(); err != nil {
// 		t.Fatal(err)
// 	}

// 	if ptrs[0].Fid != 1 || ptrs[1].Fid != 1 {
// 		t.Errorf("expected both entries in segment 1")
// 	}

// 	expectedFirstOffset := int64(8 + len(entries[0].key))

// 	if ptrs[0].Offset != expectedFirstOffset {
// 		t.Errorf(
// 			"first offset = %d, want %d",
// 			ptrs[0].Offset,
// 			expectedFirstOffset,
// 		)
// 	}

// 	expectedSecondOffset := ptrs[0].Offset + ptrs[0].Len + 8 + int64(len(entries[1].key))

// 	if ptrs[1].Offset != expectedSecondOffset {
// 		t.Errorf(
// 			"second offset = %d, want %d",
// 			ptrs[1].Offset,
// 			expectedSecondOffset,
// 		)
// 	}

// 	for i, ptr := range ptrs {
// 		result := make([]byte, len(entries[i].value))
// 		err := v.ReadValue(ptr, result)

// 		if err != nil {
// 			t.Errorf(err.Error())
// 		}

// 		if string(result) != string(entries[i].value) {
// 			t.Errorf("value = %q, want %q", result, entries[i].value)
// 		}

// 	}
// }

// func TestRotate(t *testing.T) {
// 	path := filepath.Join(t.TempDir(), "vlog")

// 	v, err := OpenVLog(path)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	defer v.Close()

// 	if err := v.rotate(); err != nil {
// 		t.Fatalf("Rotate() error = %v", err)
// 	}

// 	if v.activeFid != 2 {
// 		t.Fatalf("activeFid = %d, want 2", v.activeFid)
// 	}

// 	if v.offset != 0 {
// 		t.Fatalf("offset = %d, want 0", v.offset)
// 	}

// 	if len(v.files) != 2 {
// 		t.Fatalf("files = %d, want 2", len(v.files))
// 	}

// 	if v.files[2] == nil {
// 		t.Fatal("files[2] is nil")
// 	}
// }

// func TestReopenVLog(t *testing.T) {
// 	path := filepath.Join(t.TempDir(), "vlog")

// 	v, err := OpenVLog(path)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	entry := Entry{
// 		key:   []byte("hello"),
// 		value: []byte("world"),
// 	}

// 	ptr, err := v.Append(entry)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	if err := v.Sync(); err != nil {
// 		t.Fatal(err)
// 	}

// 	if err := v.Close(); err != nil {
// 		t.Fatal(err)
// 	}

// 	v, err = OpenVLog(path)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	defer v.Close()

// 	if v.activeFid != 1 {
// 		t.Errorf("activeFid = %d, want 1", v.activeFid)
// 	}

// 	expectedOffset := int64(8 + len(entry.key))

// 	if v.offset != expectedOffset+int64(len(entry.value)) {
// 		t.Errorf(
// 			"offset = %d, want %d",
// 			v.offset,
// 			expectedOffset+int64(len(entry.value)),
// 		)
// 	}

// 	result := make([]byte, len(entry.value))

// 	err = v.ReadValue(ptr, result)
// 	if err != nil {
// 		t.Fatalf("ReadValue() error = %v", err)
// 	}

// 	if string(result) != string(entry.value) {
// 		t.Errorf("value = %q, want %q", result, entry.value)
// 	}
// }
