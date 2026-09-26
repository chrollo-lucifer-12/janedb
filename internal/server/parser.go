package server

import (
	"bytes"
	"fmt"
)

type Request struct {
	Command string
	Key     string
	Value   string
}

func ParseRequest(buf []byte) (Request, error) {
	parts := bytes.SplitN(bytes.TrimSpace(buf), []byte(" "), 3)

	switch string(parts[0]) {
	case "SET":
		if len(parts) != 3 {
			return Request{}, fmt.Errorf("usage: SET key value")
		}

		return Request{
			Command: "SET",
			Key:     string(parts[1]),
			Value:   string(parts[2]),
		}, nil

	case "GET", "DEL":
		if len(parts) != 2 {
			return Request{}, fmt.Errorf("usage: %s key", parts[0])
		}

		return Request{
			Command: string(parts[0]),
			Key:     string(parts[1]),
		}, nil

	default:
		return Request{}, fmt.Errorf("unknown command")
	}
}
