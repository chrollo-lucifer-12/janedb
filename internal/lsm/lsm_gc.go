package lsm

import (
	"log"

	"github.com/janedb/internal/vlog"
)

type update struct {
	key []byte
	ptr vlog.ValuePointer
}

func (lsm *LSM) RunGarbageCollector() error {
	log.Println("GC: entered")

	it := lsm.v.Iterator()

	oldTail := it.GetStart()
	log.Println("GC: iterator created", oldTail)

	var updates []update

	for {
		entry, oldPtr, ok := it.Next()
		if !ok {
			break
		}

		log.Printf("GC: key=%s ptr=%+v\n", entry.Key, oldPtr)

		ptr, err := lsm.getPtr(entry.Key)
		if err != nil {
			log.Println("GC: getPtr error:", err)
			continue
		}

		if oldPtr != ptr {
			log.Println("GC: obsolete")
			continue
		}

		log.Println("GC: LIVE")

		newPtr, err := lsm.v.Append(entry)
		if err != nil {
			return err
		}

		updates = append(updates, update{
			key: append([]byte(nil), entry.Key...),
			ptr: newPtr,
		})
	}

	log.Println("GC: scan finished")

	newTail := it.GetStart()

	log.Println("GC: syncing vlog")
	if err := lsm.v.Sync(); err != nil {
		return err
	}
	log.Println("GC: vlog synced")

	for _, u := range updates {
		lsm.m.UpdatePtr(u.key, u.ptr)
	}

	log.Println("GC: saving tail")
	if err := SaveTail(newTail); err != nil {
		return err
	}

	lsm.v.SetTail(newTail)

	log.Println("GC: reclaiming")
	if err := lsm.v.Reclaim(oldTail, newTail); err != nil {
		return err
	}

	log.Println("GC: finished")
	return nil
}
