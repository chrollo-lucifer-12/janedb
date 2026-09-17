package main

import (
	"log"

	"github.com/janedb/internal/lsm"
)

func main() {
	d, err := lsm.OpenLSM()
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := d.Close(); err != nil {
			log.Fatal(err)
		}
	}()

	// Same key, two different values.
	if err := d.Put([]byte("k"), []byte("1111111111")); err != nil {
		log.Fatal(err)
	}

	if err := d.Put([]byte("k"), []byte("2222222222")); err != nil {
		log.Fatal(err)
	}

	log.Println("before GC:")

	buf := make([]byte, 10)
	if ok := d.Get([]byte("k"), buf); ok {
		log.Printf("k = %q\n", string(buf))
	}

	log.Printf("head = %d, tail = %d\n", d.VLogHead(), d.VLogTail())

	log.Println("running GC...")

	if err := d.RunGarbageCollector(); err != nil {
		log.Fatal(err)
	}

	log.Println("after GC:")

	buf = make([]byte, 10)
	if ok := d.Get([]byte("k"), buf); ok {
		log.Printf("k = %q\n", string(buf))
	}

	log.Printf("head = %d, tail = %d\n", d.VLogHead(), d.VLogTail())
}
