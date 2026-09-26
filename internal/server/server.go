package server

import (
	"bytes"
	"fmt"

	"github.com/janedb/internal/lsm"
	"golang.org/x/sys/unix"
)

type Server struct {
	db      *lsm.LSM
	clients map[uint32]*Client
}

func NewServer(db *lsm.LSM) *Server {
	return &Server{
		db:      db,
		clients: make(map[uint32]*Client),
	}
}

func (s *Server) RunServer() error {

	fd, err := unix.Socket(
		unix.AF_INET,
		unix.SOCK_STREAM,
		0,
	)
	if err != nil {
		return fmt.Errorf("socket: %w", err)
	}
	defer unix.Close(fd)

	addr := &unix.SockaddrInet4{
		Port: 8080,
		Addr: [4]byte{0, 0, 0, 0},
	}

	if err := unix.Bind(fd, addr); err != nil {
		return fmt.Errorf("bind: %w", err)
	}

	if err := unix.Listen(fd, 128); err != nil {
		return fmt.Errorf("listen: %w", err)
	}

	for {
		connFD, _, err := unix.Accept(fd)
		if err != nil {
			return fmt.Errorf("accept: %w", err)
		}

		go s.handleConnection(connFD)
	}
}

func (s *Server) handleConnection(fd int) {
	defer unix.Close(fd)

	c := NewClient(uint32(fd))

	s.clients[uint32(fd)] = c

	for {
		n, err := unix.Read(fd, c.readBuf)
		if err != nil {
			return
		}

		if n == 0 {
			return
		}

		buf := c.readBuf[:n]
		c.writeBuf = c.writeBuf[:0]

		for len(buf) > 0 {
			idx := bytes.IndexByte(buf, '\n')
			if idx == -1 {
				break
			}

			line := bytes.TrimSpace(buf[:idx])
			buf = buf[idx+1:]

			if len(line) == 0 {
				continue
			}

			req, err := ParseRequest(line)
			if err != nil {
				c.writeBuf = append(c.writeBuf, "ERROR "...)
				c.writeBuf = append(c.writeBuf, err.Error()...)
				c.writeBuf = append(c.writeBuf, '\n')
				continue
			}

			switch req.Command {
			case "SET":
				s.db.Put([]byte(req.Key), []byte(req.Value))
				c.writeBuf = append(c.writeBuf, "OK\n"...)
			case "GET":
				s.db.Get([]byte(req.Key), c.writeBuf)
			case "DEL":
				s.db.Delete([]byte(req.Key))
				c.writeBuf = append(c.writeBuf, "OK\n"...)
			}

		}

		if len(c.writeBuf) > 0 {
			_, err := unix.Write(fd, c.writeBuf)
			if err != nil {
				return
			}
		}
	}
}
