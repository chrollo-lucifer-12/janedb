package skl

import (
	"bytes"
	"math/rand/v2"

	"github.com/janedb/internal/vlog"
)

const maxLevel = 16

type SkipList struct {
	head *node
}

type node struct {
	key   []byte
	value vlog.ValuePointer

	next []*node
}

func NewSkipList() *SkipList {
	return &SkipList{
		head: &node{
			next: make([]*node, maxLevel),
		},
	}
}

func (s *SkipList) Get(key []byte) (vlog.ValuePointer, bool) {
	x := s.head

	for level := maxLevel - 1; level >= 0; level-- {
		for x.next[level] != nil && bytes.Compare(x.next[level].key, key) < 0 {
			x = x.next[level]
		}
	}

	x = x.next[0]

	if x != nil && bytes.Equal(x.key, key) {
		return x.value, true
	}

	return vlog.ValuePointer{}, false
}

func (s *SkipList) Put(key []byte, value vlog.ValuePointer) {
	update := make([]*node, maxLevel)

	x := s.head

	for level := maxLevel - 1; level >= 0; level-- {
		for x.next[level] != nil &&
			bytes.Compare(x.next[level].key, key) < 0 {
			x = x.next[level]
		}

		update[level] = x
	}

	if x.next[0] != nil &&
		bytes.Equal(x.next[0].key, key) {

		x.next[0].value = value
		return
	}

	level := randomLevel()

	n := &node{
		key:   key,
		value: value,
		next:  make([]*node, level),
	}

	for i := 0; i < level; i++ {
		n.next[i] = update[i].next[i]
		update[i].next[i] = n
	}
}

type Iterator struct {
	skl  *SkipList
	node *node
}

func (s *SkipList) Iterator() *Iterator {
	return &Iterator{skl: s}
}

func (it *Iterator) Rewind() {
	it.node = it.skl.head.next[0]
}

func (it *Iterator) Valid() bool {
	return it.node != nil
}

func (it *Iterator) Key() []byte {
	return it.node.key
}

func (it *Iterator) Value() vlog.ValuePointer {
	return it.node.value
}

func (it *Iterator) Next() {
	if it.node != nil {
		it.node = it.node.next[0]
	}
}

func (it *Iterator) Seek(key []byte) {
	x := it.skl.head

	for level := maxLevel - 1; level >= 0; level-- {
		for x.next[level] != nil && bytes.Compare(x.next[level].key, key) < 0 {
			x = x.next[level]
		}
	}

	it.node = x.next[0]
}

func randomLevel() int {
	level := 1

	for level < maxLevel && rand.Float64() < 0.5 {
		level++
	}

	return level
}
