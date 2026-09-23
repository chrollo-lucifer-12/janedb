package lsm

import (
	"log"
	"time"

	"github.com/janedb/internal/vlog"
)

type update struct {
	key []byte
	ptr vlog.ValuePointer
}

func (lsm *LSM) RunGarbageCollector() error {

	it := lsm.v.Iterator()

	oldTail := it.GetStart()

	var updates []update

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

	newTail := it.GetStart()

	if err := lsm.v.Sync(); err != nil {
		return err
	}

	for _, u := range updates {
		lsm.m.UpdatePtr(u.key, u.ptr)
	}

	if err := saveTail(newTail); err != nil {
		return err
	}

	lsm.v.SetTail(newTail)

	if err := lsm.v.Reclaim(oldTail, newTail); err != nil {
		return err
	}

	return nil
}

func (lsm *LSM) startGC() {
	lsm.gcStop = make(chan struct{})
	lsm.gcDone = make(chan struct{})

	go func() {
		defer close(lsm.gcDone)

		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if lsm.v.GetHead()-lsm.v.GetTail() >= GCThreshold {
					if err := lsm.RunGarbageCollector(); err != nil {
						log.Printf("gc: %v", err)
					}
				}

			case <-lsm.gcStop:
				return
			}
		}
	}()
}
