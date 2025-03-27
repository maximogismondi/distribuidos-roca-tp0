package communication

import (
	"fmt"
	"strings"
)

const HEADER_SEPARATOR = ':'
const WINNERS_SEPARATOR = ','

const WINNERS_MESSAGE = "winners"
const NOT_READY_MESSAGE = "not_ready"
const SUCCESS_MESSAGE = "success"
const FAILURE_MESSAGE = "failure"

func DecodeMessage(message string) (string, string, error) {
	fields := strings.Split(message, string(HEADER_SEPARATOR))

	if len(fields) != 2 {
		return "", "", fmt.Errorf("invalid message format")
	}

	headers := []string{WINNERS_MESSAGE, NOT_READY_MESSAGE, SUCCESS_MESSAGE, FAILURE_MESSAGE}

	for _, header := range headers {
		if fields[0] == header {
			return fields[0], fields[1], nil
		}
	}

	return "", "", fmt.Errorf("invalid message type")
}

func DecodeWinnersMessage(message string) []string {
	return strings.Split(message, string(WINNERS_SEPARATOR))
}
