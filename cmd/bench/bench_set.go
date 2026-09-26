package main

import (
	"fmt"
	"io"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

const (
	clients  = 10
	duration = 30 * time.Second
)

func main() {
	var total atomic.Uint64
	var wg sync.WaitGroup

	start := time.Now()
	end := start.Add(duration)

	for i := 0; i < clients; i++ {
		wg.Add(1)

		go func(id int) {
			defer wg.Done()

			conn, err := net.Dial("tcp", "127.0.0.1:8080")
			if err != nil {
				panic(err)
			}
			defer conn.Close()

			req := []byte("SET key value\n")
			buf := make([]byte, 3)

			for time.Now().Before(end) {
				_, err := conn.Write(req)
				if err != nil {
					return
				}

				_, err = io.ReadFull(conn, buf)
				if err != nil {
					return
				}

				total.Add(1)
			}
		}(i)
	}

	wg.Wait()

	elapsed := time.Since(start)
	requests := total.Load()

	fmt.Printf("clients:   %d\n", clients)
	fmt.Printf("duration:  %v\n", elapsed)
	fmt.Printf("requests:  %d\n", requests)
	fmt.Printf("throughput: %.0f req/s\n",
		float64(requests)/elapsed.Seconds())
}
