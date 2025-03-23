package common

import (
	"bufio"
	"net"
)

const COMMUNICATION_DELIMITER = '\n'

type Socket struct {
	conn   net.Conn
	reader *bufio.Reader
}

func NewSocket(c net.Conn) Socket {
	return Socket{
		conn:   c,
		reader: bufio.NewReader(c),
	}
}

func (s *Socket) Read() (string, error) {
	message, err := s.reader.ReadString(COMMUNICATION_DELIMITER)
	if err != nil {
		return "", err
	}

	return message, nil
}

func (s *Socket) Write(message string) error {
	message_bytes := []byte(message + string(COMMUNICATION_DELIMITER))
	bytes_written := 0

	for bytes_written < len(message_bytes) {
		n, err := s.conn.Write(message_bytes[bytes_written:])
		if err != nil {

			return err
		}
		bytes_written += n
	}

	return nil
}

func (s *Socket) Close() error {
	return s.conn.Close()
}
