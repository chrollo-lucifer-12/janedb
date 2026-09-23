package vlog

type Iterator struct {
	vlog  *VLog
	start int64
	end   int64
}

func (v *VLog) Iterator() *Iterator {

	v.headmu.Lock()
	head := v.head
	v.headmu.Unlock()

	v.tailmu.Lock()
	tail := v.tail
	v.tailmu.Unlock()

	return &Iterator{
		vlog:  v,
		start: tail,
		end:   head,
	}
}

func (it *Iterator) Next() (Entry, ValuePointer, bool) {
	if it.start >= it.end {
		return Entry{}, ValuePointer{}, false
	}

	entry, err := DecodeEntry(it.vlog.data[it.start:it.end])
	if err != nil {
		return Entry{}, ValuePointer{}, false
	}

	ptr := ValuePointer{
		Offset: it.start + 8 + int64(len(entry.Key)),
		Len:    int64(len(entry.Value)),
	}

	it.start += int64(
		8 + len(entry.Key) + len(entry.Value),
	)

	return entry, ptr, true
}

func (it *Iterator) GetStart() int64 {
	return it.start
}
