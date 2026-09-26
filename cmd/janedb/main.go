package main

import (
	"fmt"

	"github.com/janedb/internal/lsm"
)

func main() {
	db, err := lsm.OpenLSM()
	if err != nil {
		panic(err)
	}
	defer db.Close()

	buf := make([]byte, 1024)

	ok := db.Get([]byte("k"), buf)
	if !ok {
		fmt.Println("not found")
	} else {
		fmt.Println(string(buf))
	}

	// s := server.NewServer(db)

	// if err := s.RunServer(); err != nil {
	// 	panic(err)
	// }
}
