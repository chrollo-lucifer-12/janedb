package vlog

import "fmt"

func (v *VLog) Append(entry Entry) (ValuePointer, error) {

	v.headmu.Lock()
	defer v.headmu.Unlock()

	size := 8 + len(entry.Key) + len(entry.Value) + 8 + 1

	if v.head+int64(size) > int64(len(v.data)) {
		return ValuePointer{}, fmt.Errorf("vlog full")
	}

	n := EncodeEntry(entry, v.data[v.head:v.head+int64(size)])

	ptr := ValuePointer{
		Offset: v.head + 8 + int64(len(entry.Key)),
		Len:    int64(len(entry.Value)),
	}

	v.head += int64(n)

	return ptr, nil
}
