package communication

import (
	"bufio"
	"net"
)

const COMMUNICATION_DELIMITER = '\n'

type ServerSocket struct {
	conn   net.Conn
	reader *bufio.Reader
}

func Connect(address string, agencyId string) (ServerSocket, error) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return ServerSocket{}, err
	}

	serverSocket := ServerSocket{
		conn:   conn,
		reader: bufio.NewReader(conn),
	}

	startMessage := EncodedIdentificationMessage(agencyId)
	err = serverSocket.Write(startMessage)

	if err != nil {
		conn.Close()
		return ServerSocket{}, err
	}

	return serverSocket, nil
}

func (s *ServerSocket) Read() (string, error) {
	message, err := s.reader.ReadString(COMMUNICATION_DELIMITER)
	if err != nil {
		return "", err
	}

	return message[:len(message)-1], nil
}

func (s *ServerSocket) Write(message string) error {
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

func (s *ServerSocket) Close() error {
	return s.conn.Close()
}
