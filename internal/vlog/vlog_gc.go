package vlog

type Iterator struct {
	vlog   *VLog
	offset int64
	end    int64
}

func (v *VLog) Iterator() *Iterator {
	return &Iterator{
		vlog:   v,
		offset: v.tail,
		end:    v.head,
	}
}

func (it *Iterator) Next() (Entry, ValuePointer, bool) {
	if it.offset >= it.end {
		return Entry{}, ValuePointer{}, false
	}

	entry, err := DecodeEntry(it.vlog.data[it.offset:])
	if err != nil {
		return Entry{}, ValuePointer{}, false
	}

	ptr := ValuePointer{
		Offset: it.offset + 8 + int64(len(entry.Key)),
		Len:    int64(len(entry.Value)),
	}

	it.offset += int64(
		8 + len(entry.Key) + len(entry.Value),
	)

	return entry, ptr, true
}
