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
	defer d.Close()

	buf := make([]byte, 10)
	d.Get([]byte("k"), buf)

	log.Println(string(buf))
}
