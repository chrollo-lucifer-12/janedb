package vlog

func (v *VLog) Append(entry Entry) (ValuePointer, error) {

	v.headmu.Lock()
	defer v.headmu.Unlock()

	EncodeEntry(entry, v.data[v.head:])

	ptr := ValuePointer{
		Offset: v.head + 8 + int64(len(entry.Key)),
		Len:    int64(len(entry.Value)),
	}

	v.head += int64(8 + len(entry.Key) + len(entry.Value))

	return ptr, nil
}
