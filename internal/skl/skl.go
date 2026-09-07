package skl

import (
	"bytes"
	"math/rand/v2"

	"github.com/janedb/internal/vlog"
)

const MaxLevel = 16
const P = 0.5
const invalidOffset = 0

type node struct {
	keyOffset uint32
	keySize   uint32

	value vlog.ValuePointer

	towerOffset uint32

	height uint8
}

type Skiplist struct {
	head  uint32
	level int
	arena *Arena
}

func NewSkiplist() *Skiplist {

	a := NewArena(64 << 20)
	headOffset, _ := a.allocateNode(nil, vlog.ValuePointer{}, MaxLevel)

	return &Skiplist{
		level: 0,
		arena: a,
		head:  headOffset,
	}
}

func (skl *Skiplist) Search(key []byte) (vlog.ValuePointer, bool) {
	curr := skl.arena.getNode(skl.head)

	for i := skl.level; i >= 0; i-- {
		for {

			nextOffset := skl.arena.getNext(curr.towerOffset, i)

			if nextOffset == invalidOffset {
				break
			}

			next := skl.arena.getNode(nextOffset)

			nextKey := skl.arena.getBytes(next.keyOffset, next.keySize)

			if bytes.Compare(nextKey, key) >= 0 {
				break
			}

			curr = next
		}
	}

	nextOffset := skl.arena.getNext(curr.towerOffset, 0)

	if nextOffset == invalidOffset {
		return vlog.ValuePointer{}, false
	}

	next := skl.arena.getNode(nextOffset)

	nextKey := skl.arena.getBytes(next.keyOffset, next.keySize)

	if bytes.Equal(nextKey, key) {
		return next.value, true
	}

	return vlog.ValuePointer{}, false
}

func (skl *Skiplist) Insert(key []byte, value vlog.ValuePointer) {

	update := make([]uint32, MaxLevel+1)

	currOffset := skl.head
	curr := skl.arena.getNode(skl.head)

	for i := skl.level; i >= 0; i-- {
		for {

			nextOffset := skl.arena.getNext(curr.towerOffset, i)

			if nextOffset == invalidOffset {
				break
			}

			next := skl.arena.getNode(nextOffset)
			nextKey := skl.arena.getBytes(next.keyOffset, next.keySize)

			if bytes.Compare(nextKey, key) >= 0 {
				break
			}

			currOffset = nextOffset
			curr = next
		}

		update[i] = currOffset
	}

	nextOffset := skl.arena.getNext(curr.towerOffset, 0)

	if nextOffset != 0 {
		next := skl.arena.getNode(nextOffset)
		nextKey := skl.arena.getBytes(next.keyOffset, next.keySize)

		if bytes.Equal(nextKey, key) {
			next.value = value
			return
		}
	}

	rlvl := randomLevel()
	if int(rlvl) > skl.level {
		for i := skl.level + 1; i <= int(rlvl); i++ {
			update[i] = skl.head
		}
		skl.level = int(rlvl)
	}

	newOffset, newNode := skl.arena.allocateNode(key, value, rlvl)

	for i := 0; i <= int(rlvl); i++ {
		prev := skl.arena.getNode(update[i])

		nextOffset := skl.arena.getNext(prev.towerOffset, i)

		skl.arena.setNext(newNode.towerOffset, i, nextOffset)
		skl.arena.setNext(prev.towerOffset, i, newOffset)
	}
}

func (skl *Skiplist) GetSize() uint64 {
	return uint64(skl.arena.n)
}

type Iterator struct {
	skl   *Skiplist
	node  uint32
	Key   []byte
	Value vlog.ValuePointer
}

func GetIterator(skl *Skiplist) *Iterator {
	return &Iterator{skl: skl, node: skl.head}
}

func (i *Iterator) GetNext() bool {

	if i.node == 0 {
		return false
	}

	node := i.skl.arena.getNode(i.node)

	i.Key = i.skl.arena.getBytes(node.keyOffset, node.keySize)
	i.Value = node.value

	i.node = i.skl.arena.getNext(i.node, 0)

	return true
}

func randomLevel() uint8 {
	lvl := uint8(0)
	for rand.Float64() < P && lvl < MaxLevel {
		lvl++
	}
	return lvl
}
