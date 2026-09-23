package vlog

type RecoverEntry struct {
	Key      []byte
	Ptr      ValuePointer
	Sequence uint64
	VType    uint8
}

func (v *VLog) Recover() map[string]RecoverEntry {

	values := make(map[string]RecoverEntry)

	start := int64(0)

	for start < int64(len(v.data)) {

		entry, err := DecodeEntry(v.data[start:])
		if err != nil {
			break
		}

		valueOffset := start + 8 + int64(len(entry.Key))

		values[string(entry.Key)] = RecoverEntry{
			Key: []byte(string(entry.Key)),
			Ptr: ValuePointer{
				Offset: valueOffset,
				Len:    int64(len(entry.Value)),
			},
			Sequence: entry.Sequence,
			VType:    entry.VType,
		}

		start += 8 + int64(len(entry.Key)) + int64(len(entry.Value)) + 9
	}

	return values
}
