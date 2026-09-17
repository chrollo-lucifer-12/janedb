package lsm

import (
	"github.com/janedb/internal/vlog"
)

func (lsm *LSM) RunGarbageCollector() {

	it := lsm.v.Iterator()

	for {

		entry, vptr, ok := it.Next()
		if !ok {
			break
		}

		ptr, err := lsm.getPtr(entry.Key)
		if err == nil {
			if vptr == ptr {
				if err := lsm.appendtoHead(entry); err != nil {
					return
				}
			}
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
