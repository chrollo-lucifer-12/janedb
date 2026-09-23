package bloom

import (
	"encoding/binary"
	"io"
	"os"

	"github.com/spaolacci/murmur3"
)

const (
	bitsPerKey = 10
	numProbes  = 7
)

type BF struct {
	bits []byte
	k    uint8
}

func NewBloomFilter(numKeys int) *BF {

	numBits := numKeys * bitsPerKey

	numBytes := (numBits + 7) / 8
	numBits = numBytes * 8

	return &BF{
		bits: make([]byte, numBytes),
		k:    numProbes,
	}
}

func (b *BF) Add(key []byte) {
	h1, h2 := hashes(key)

	numBits := uint64(len(b.bits) * 8)

	for i := uint64(0); i < uint64(b.k); i++ {
		pos := (h1 + i*h2) % numBits

		byteIndex := pos >> 3
		bitIndex := pos & 7

		b.bits[byteIndex] |= 1 << bitIndex
	}
}

func (b *BF) MayContain(key []byte) bool {
	h1, h2 := hashes(key)

	numBits := uint64(len(b.bits) * 8)

	for i := uint64(0); i < uint64(b.k); i++ {
		pos := (h1 + i*h2) % numBits

		byteIndex := pos >> 3
		bitIndex := pos & 7

		if b.bits[byteIndex]&(1<<bitIndex) == 0 {
			return false
		}
	}

	return true
}

func (b *BF) Marshal(file *os.File) error {
	var header [5]byte

	header[0] = b.k
	binary.LittleEndian.PutUint32(header[1:], uint32(len(b.bits)*8))

	if _, err := file.Write(header[:]); err != nil {
		return err
	}

	if _, err := file.Write(b.bits); err != nil {
		return err
	}

	return nil
}

func (b *BF) Unmarshal(file *os.File) error {
	var header [5]byte

	if _, err := io.ReadFull(file, header[:]); err != nil {
		return err
	}

	b.k = header[0]

	numBits := binary.LittleEndian.Uint32(header[1:])
	numBytes := (numBits + 7) / 8

	b.bits = make([]byte, numBytes)

	if _, err := io.ReadFull(file, b.bits); err != nil {
		return err
	}

	return nil
}

func hashes(key []byte) (uint64, uint64) {
	return murmur3.Sum128(key)
}
