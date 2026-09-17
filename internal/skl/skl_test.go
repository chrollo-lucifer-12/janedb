package skl

// func TestSkiplistInsertSearch(t *testing.T) {
// 	skl := NewSkiplist()

// 	ptr1 := vlog.ValuePointer{Fid: 1, Offset: 100, Len: 10}
// 	ptr2 := vlog.ValuePointer{Fid: 1, Offset: 200, Len: 20}

// 	skl.Insert([]byte("apple"), ptr1)
// 	skl.Insert([]byte("banana"), ptr2)

// 	got, _ := skl.Search([]byte("apple"))
// 	// if !ok {
// 	// 	t.Fatal("expected apple to be found")
// 	// }

// 	if got != ptr1 {
// 		t.Fatalf("expected %+v, got %+v", ptr1, got)
// 	}

// 	got, _ = skl.Search([]byte("banana"))
// 	// if !ok {
// 	// 	t.Fatal("expected banana to be found")
// 	// }

// 	if got != ptr2 {
// 		t.Fatalf("expected %+v, got %+v", ptr2, got)
// 	}
// }

// func TestSkiplistUpdateAndMissing(t *testing.T) {
// 	skl := NewSkiplist()

// 	oldPtr := vlog.ValuePointer{Fid: 1, Offset: 100, Len: 10}
// 	newPtr := vlog.ValuePointer{Fid: 2, Offset: 500, Len: 50}

// 	skl.Insert([]byte("key"), oldPtr)
// 	skl.Insert([]byte("key"), newPtr)

// 	got, _ := skl.Search([]byte("key"))
// 	// if !ok {
// 	// 	t.Fatal("expected key to be found")
// 	// }

// 	if got != newPtr {
// 		t.Fatalf("expected updated pointer %+v, got %+v", newPtr, got)
// 	}

// 	_, ok := skl.Search([]byte("missing"))
// 	if ok {
// 		t.Fatal("expected missing key to not be found")
// 	}
// }

// func BenchmarkSkiplistInsert(b *testing.B) {
// 	b.StopTimer()

// 	skl := NewSkiplist()

// 	keys := make([][]byte, 100000)
// 	values := make([]vlog.ValuePointer, 100000)

// 	for i := 0; i < 100000; i++ {
// 		keys[i] = []byte("key" + strconv.Itoa(i))
// 		values[i] = vlog.ValuePointer{
// 			Fid:    1,
// 			Offset: int64(i),
// 			Len:    100,
// 		}
// 	}

// 	b.StartTimer()

// 	for n := 0; n < b.N; n++ {
// 		for i := 0; i < 100000; i++ {
// 			skl.Insert(keys[i], values[i])
// 		}
// 	}
// }

// func BenchmarkSkiplistSearch(b *testing.B) {
// 	skl := NewSkiplist()

// 	for i := 0; i < 1000; i++ {
// 		key := []byte("key" + string(rune(i)))
// 		value := vlog.ValuePointer{
// 			Fid:    1,
// 			Offset: int64(i),
// 			Len:    100,
// 		}

// 		skl.Insert(key, value)
// 	}

// 	keys := [][]byte{
// 		[]byte("key100"),
// 		[]byte("key500"),
// 		[]byte("key900"),
// 	}

// 	b.ResetTimer()

// 	for i := 0; i < b.N; i++ {
// 		skl.Search(keys[i%len(keys)])
// 	}

// }
