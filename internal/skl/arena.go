package skl

import (
	"encoding/binary"
	"unsafe"

	"github.com/janedb/internal/vlog"
)

type Arena struct {
	buf []byte
	n   uint32
}

func NewArena(size uint32) *Arena {
	return &Arena{
		buf: make([]byte, size),
		n:   8,
	}
}

func (a *Arena) allocate(size uint32) uint32 {
	offset := a.n
	a.n += size

	if a.n > uint32(len(a.buf)) {
		panic("arena full")
	}

	return offset
}

func (a *Arena) getBytes(offset, size uint32) []byte {
	return a.buf[offset : offset+size]
}

func (a *Arena) allocateKey(key []byte) (uint32, uint32) {

	keyLen := uint32(len(key))

	offset := a.allocate(keyLen)
	if keyLen > 0 {
		copy(a.getBytes(offset, keyLen), key)
	}

	return offset, keyLen
}

func (a *Arena) allocateNode(key []byte, value vlog.ValuePointer, height uint8) (uint32, *node) {

	keyOffset, keySize := a.allocateKey(key)

	size := uint32(unsafe.Sizeof(node{}))
	offset := a.allocate(size)
	n := (*node)(unsafe.Pointer(&a.buf[offset]))

	n.keyOffset = keyOffset
	n.keySize = keySize
	n.value = value
	n.height = height

	towerSize := uint32((height + 1) * 4)
	n.towerOffset = a.allocate(towerSize)

	return offset, n
}

func (a *Arena) getNode(offset uint32) *node {
	return (*node)(unsafe.Pointer(&a.buf[offset]))
}

func (a *Arena) getNext(towerOffset uint32, level int) uint32 {
	offset := towerOffset + uint32(level*4)
	return binary.LittleEndian.Uint32(a.buf[offset : offset+4])
}

func (a *Arena) setNext(towerOffset uint32, level int, value uint32) {
	offset := towerOffset + uint32(level*4)
	binary.LittleEndian.PutUint32(a.buf[offset:offset+4], value)

}
