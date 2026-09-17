package vlog

import "github.com/edsrzf/mmap-go"

func (lf *logFile) GetData() mmap.MMap { return lf.data }

func (v *VLog) GetFiles() map[uint32]*logFile {
	return v.files
}

func (v *VLog) GetActiveFID() uint32 {
	return v.activeFid
}

func Upperbound(data []byte) (int64, error) {
	return findOffset(data)
}
