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

}
