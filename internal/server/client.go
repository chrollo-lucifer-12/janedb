package server

type Client struct {
	fd uint32

	readBuf  []byte
	writeBuf []byte
	valueBuf []byte
}

func NewClient(fd uint32) *Client {
	return &Client{
		fd:       fd,
		readBuf:  make([]byte, 2048),
		writeBuf: make([]byte, 2048),
		valueBuf: make([]byte, 64*1024),
	}
}
