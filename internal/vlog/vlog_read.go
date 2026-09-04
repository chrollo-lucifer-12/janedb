package vlog

import (
	"fmt"
	"io"
	"sync"
)

func (v *VLog) ReadValue(ptr ValuePointer, buf []byte) error {
	v.mu.RLock()

	file, ok := v.files[ptr.Fid]
	if !ok {
		return fmt.Errorf("read vlog: file not found for fid %d", ptr.Fid)
	}

	v.mu.RUnlock()

	n, err := file.ReadAt(buf, ptr.Offset)
	if err != nil {
		return fmt.Errorf("read vlog: %w", err)
	}

	if int64(n) != ptr.Len {
		return io.ErrUnexpectedEOF
	}

	return nil
}

func (v *VLog) Read(ptrs []ValuePointer, results []ReadResult) {
	var wg sync.WaitGroup
	wg.Add(len(ptrs))

	for i, ptr := range ptrs {

		results[i].Ptr = ptr

		v.readQueue <- readJob{
			index:   i,
			ptr:     ptr,
			results: results,
			wg:      &wg,
		}
	}

	wg.Wait()

}

func (v *VLog) readWorker() {
	for job := range v.readQueue {
		result := &job.results[job.index]

		result.Ptr = job.ptr

		if cap(result.Value) < int(job.ptr.Len) {
			result.Value = make([]byte, job.ptr.Len)
		} else {
			result.Value = result.Value[:job.ptr.Len]
		}

		if err := v.ReadValue(job.ptr, result.Value); err != nil {
			result.Err = err
		}

		job.wg.Done()
	}
}
