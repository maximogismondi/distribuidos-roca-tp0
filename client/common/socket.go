package common

import (
	"bufio"
	"net"
)

const COMMUNICATION_DELIMITER = '\n'
const AGENCY_IDENTIFICATION_DELIMITER = ':'
const AGENCY_IDENTIFICATION_MESSAGE = "agency"

type Socket struct {
	conn   net.Conn
	reader *bufio.Reader
}

func NewSocket(c net.Conn, agencyId string) (Socket, error) {

	socket := Socket{
		conn:   c,
		reader: bufio.NewReader(c),
	}

	startMessage := AGENCY_IDENTIFICATION_MESSAGE + string(AGENCY_IDENTIFICATION_DELIMITER) + agencyId
	err := socket.Write(startMessage)

	if err != nil {
		return socket, err
	}

	return socket, nil
}

func (s *Socket) Read() (string, error) {
	message, err := s.reader.ReadString(COMMUNICATION_DELIMITER)
	if err != nil {
		return "", err
	}

	return message[:len(message)-1], nil
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
