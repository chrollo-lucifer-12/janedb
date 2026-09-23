package vlog

func (v *VLog) Recover() map[string]ValuePointer {

	values := make(map[string]ValuePointer)

	start := int64(0)

	for start < int64(len(v.data)) {

		entry, err := DecodeEntry(v.data[start:])
		if err != nil {
			break
		}

		if entry.Value == nil {
			delete(values, string(entry.Key))
		} else {
			values[string(entry.Key)] = ValuePointer{
				Offset: start + 8 + int64(len(entry.Key)),
				Len:    int64(len(entry.Value)),
			}
		}

		start += 8 + int64(len(entry.Key)) + int64(len(entry.Value))
	}

	return values
}
