package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/janedb/internal/lsm"
)

func main() {

	d, err := lsm.OpenLSM()
	if err != nil {
		panic(err)
	}

	r := bufio.NewReader(os.Stdin)

	for {
		data, err := r.ReadString('\n')

		if err != nil {
			panic(err)
		}

		cmd := strings.Fields(data)

		switch cmd[0] {
		case "put":
			key := cmd[1]
			val := cmd[2]

			if err := d.Put([]byte(key), []byte(val)); err != nil {
				panic(err)
			}

			fmt.Printf("put for key: %s\n", key)

		case "get":

			key := cmd[1]
			buf := make([]byte, 1024)
			d.Get([]byte(key), buf)

			fmt.Println(string(buf))

		default:
			break
		}
	}

}
