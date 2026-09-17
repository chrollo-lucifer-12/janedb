package lsm

import (
	"github.com/janedb/internal/vlog"
)

type update struct {
	key []byte
	ptr vlog.ValuePointer
}

func (lsm *LSM) RunGarbageCollector() error {

	it := lsm.v.Iterator()
	var updates []update

	oldTail := it.GetStart()

	for {

		entry, oldPtr, ok := it.Next()
		if !ok {
			break
		}

		ptr, err := lsm.getPtr(entry.Key)
		if err != nil {
			continue
		}

		if oldPtr != ptr {
			continue
		}

		newPtr, err := lsm.v.Append(entry)
		if err != nil {
			return err
		}

		updates = append(updates, update{
			key: append([]byte(nil), entry.Key...),
			ptr: newPtr,
		})
	}

	if err := lsm.v.Sync(); err != nil {
		return err
	}

	for _, u := range updates {
		lsm.m.UpdatePtr(u.key, u.ptr)
	}

	meta, err := lsm.m.Flush()
	if err != nil {
		return err
	}

	lsm.level = append(lsm.level, meta)

	lsm.v.SetTail(it.GetStart())

	return lsm.v.Reclaim(oldTail, it.GetStart())
}
