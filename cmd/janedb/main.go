package main

import (
	"github.com/janedb/internal/lsm"
	"github.com/janedb/internal/server"
)

func main() {
	db, err := lsm.OpenLSM()
	if err != nil {
		panic(err)
	}
	defer db.Close()

	s := server.NewServer(db)

	if err := s.RunServer(); err != nil {
		panic(err)
	}
}
