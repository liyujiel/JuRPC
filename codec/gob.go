package codec

import (
	"bufio"
	"encoding/gob"
	"io"
	"log"
)

type GobCodec struct {
	Codec
	connect io.ReadWriteCloser
	buf     *bufio.Writer
	decoder *gob.Decoder
	econder *gob.Encoder
}

func NewGobCodec(conn io.ReadWriteCloser) Codec {
	buf := bufio.NewWriter(conn)
	return &GobCodec{
		connect: conn,
		buf:     buf,
		decoder: gob.NewDecoder(conn),
		econder: gob.NewEncoder(buf),
	}
}

func (c *GobCodec) ReadHeader(h *Header) error {
	return c.decoder.Decode(h)
}

func (c *GobCodec) ReadBody(body interface{}) error {
	return c.decoder.Decode(body)
}

func (c *GobCodec) Write(h *Header, body interface{}) error {
	defer func() {
		err := c.buf.Flush()
		if err != nil {
			c.Close()
		}
	}()

	if err := c.econder.Encode(h); err != nil {
		log.Println("rpc codec: gob error encoding header:", err)
		return err
	}

	if err := c.econder.Encode(body); err != nil {
		log.Println("rpc codec: gob error encoding body:", err)
		return err
	}

	return nil
}

func (c *GobCodec) Close() error {
	return c.connect.Close()
}
