package lsm

import (
	"github.com/janedb/internal/vlog"
)

func (lsm *LSM) RunGarbageCollector() {

	for fid, file := range lsm.v.GetFiles() {
		if fid == lsm.v.GetActiveFID() {
			continue
		}

		start := int64(0)

		data := file.GetData()

		limit, err := vlog.Upperbound(data)
		if err != nil {
			continue
		}

		for start < limit {
			entry, err := vlog.DecodeEntry(data[start:])
			if err != nil {
				break
			}

			ptr, err := lsm.getPtr(entry.Key)
			if err == nil {
				recordPtr := vlog.ValuePointer{
					Fid:    fid,
					Offset: start + 8 + int64(len(entry.Key)),
					Len:    int64(len(entry.Value)),
				}

				if recordPtr == ptr {
					if err := lsm.appendtoHead(entry); err != nil {
						return
					}
				}
			}

			start += 8 + int64(len(entry.Key)) + int64(len(entry.Value))
		}
	}
}

func (lsm *LSM) appendtoHead(entry vlog.Entry) error {
	ptr, err := lsm.v.Append(entry)
	if err != nil {
		return err
	}

	lsm.m.UpdatePtr(entry.Key, ptr)

	return nil
}
