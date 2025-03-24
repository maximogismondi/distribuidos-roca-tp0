package common

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/op/go-logging"
)

const BATCH_SEPARATOR = '*'
const SUCCESS_MESSAGE = "success"

const MAX_BATCH_BYTES = 8*1024 - 1 // 8kB - 1 (comunicator delimiter)

var log = logging.MustGetLogger("log")

// ClientConfig Configuration used by the client
type ClientConfig struct {
	ID            string
	ServerAddress string
	BatchAmount   int
	DataFilePath  string
}

// Client Entity that encapsulates how
type Client struct {
	config     ClientConfig
	socket     Socket
	betChannel chan Bet
	nextBet    string
	running    bool
}

// NewClient Initializes a new client receiving the configuration
// as a parameter
func NewClient(config ClientConfig) *Client {
	client := &Client{
		config:     config,
		betChannel: make(chan Bet, config.BatchAmount),
		running:    true,
	}
	return client
}

// CreateClientSocket Initializes client socket. In case of
// failure, error is printed in stdout/stderr and exit 1
// is returned
func (c *Client) createClientSocket() error {
	conn, err := net.Dial("tcp", c.config.ServerAddress)
	if err != nil {
		log.Criticalf(
			"action: connect | result: fail | client_id: %v | error: %v",
			c.config.ID,
			err,
		)
	}
	c.socket = NewSocket(conn)
	return nil
}

func (c *Client) dataToBets(filePath string) {
	file, err := os.Open(filePath)
	if err != nil {
		close(c.betChannel)
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		bet, err := FromCSVLine(line)
		if err != nil {
			continue
		}
		c.betChannel <- bet
	}
	close(c.betChannel)
}

func (c *Client) buildBatch() string {
	batch := []string{}
	accumlatedBytes := 0

	if c.nextBet != "" {
		batch = append(batch, c.nextBet)
		accumlatedBytes += len(c.nextBet)
		c.nextBet = ""
	}

	for len(batch) < c.config.BatchAmount {
		bet, ok := <-c.betChannel
		if !ok {
			break
		}
		betStr := bet.String()
		byteLength := len(betStr)

		if accumlatedBytes+byteLength+1 > MAX_BATCH_BYTES {
			c.nextBet = betStr
			break
		}

		if len(batch) > 0 {
			accumlatedBytes++ // For the separator
		}

		batch = append(batch, betStr)
		accumlatedBytes += byteLength
	}

	return strings.Join(batch, string(BATCH_SEPARATOR))
}

func (c *Client) sendBatch(batch string) error {
	if err := c.socket.Write(batch); err != nil {
		log.Criticalf("action: apuesta_enviada | result: fail | cantidad: %v", len(batch))
		return err
	}

	msg, err := c.socket.Read()
	if err != nil || msg != SUCCESS_MESSAGE {
		log.Criticalf("action: apuesta_enviada | result: fail | cantidad: %v", len(batch))
		return fmt.Errorf("send failed or bad response")
	}

	log.Infof("action: apuesta_enviada | result: success | cantidad: %v", len(batch))
	return nil
}

// StartClientLoop Send messages to the client until some time threshold is met
func (c *Client) StartClientLoop() {
	// Create the connection the server in every loop iteration.
	err := c.createClientSocket()
	if err != nil {
		return
	}

	defer c.cleanUp()

	// Go routine to read the data from the file
	go c.dataToBets(c.config.DataFilePath)

	// Build and send batches until the channel is closed (no more bets)
	for c.running {
		batch := c.buildBatch()

		if len(batch) == 0 {
			break
		}

		err := c.sendBatch(batch)
		if err != nil {
			break
		}
	}
}

func (c *Client) StopClientLoop() {
	c.running = false
}

func (c *Client) cleanUp() {
	c.socket.Close()
	close(c.betChannel)
	log.Infof("action: exit | result: success | client_id: %v", c.config.ID)
}
