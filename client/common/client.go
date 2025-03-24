package common

import (
	"bufio"
	"net"
	"os"
	"strings"

	"github.com/op/go-logging"
)

const BATCH_SEPARATOR = '*'
const FINISH_MESSAGE = "finish"
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
	config   ClientConfig
	socket   Socket
	bets     chan Bet
	freeBets chan struct{}
	done     chan struct{}
	nextBet  string
	running  bool
}

// NewClient Initializes a new client receiving the configuration
// as a parameter
func NewClient(config ClientConfig) *Client {
	client := &Client{
		config:   config,
		bets:     make(chan Bet, config.BatchAmount),
		freeBets: make(chan struct{}, config.BatchAmount),
		done:     make(chan struct{}),
		running:  true,
	}

	// Fill the freeBets channel with the amount of bets
	for i := 0; i < config.BatchAmount; i++ {
		client.freeBets <- struct{}{}
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

func (c *Client) readBets() {
	file, err := os.Open(c.config.DataFilePath)

	if err != nil {
		close(c.bets)
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() && c.running {
		line := scanner.Text()
		bet, err := FromCSVLine(line, c.config.ID)
		if err != nil {
			log.Errorf("Error parsing line: %v", line)
			continue
		}

		select {
		case <-c.freeBets:
			c.bets <- bet
		case <-c.done:
			close(c.bets)
			return
		}
	}
	close(c.bets)
}

func (c *Client) buildBatch() []string {
	batch := []string{}
	accumlatedBytes := 0

	if c.nextBet != "" {
		batch = append(batch, c.nextBet)
		accumlatedBytes += len(c.nextBet)
		c.nextBet = ""
	}

	for len(batch) < c.config.BatchAmount {
		bet, ok := <-c.bets
		if !ok {
			break
		}

		c.freeBets <- struct{}{}
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

	return batch
}

func (c *Client) sendBatch(batch []string) error {
	log.Debugf("Sending batch: %v", batch)

	batchStr := strings.Join(batch, string(BATCH_SEPARATOR))

	if err := c.socket.Write(batchStr); err != nil {
		log.Criticalf("action: apuesta_enviada | result: fail | cantidad: %v", len(batch))
		return err
	}

	return nil
}

func (c *Client) waitForServerResponse() (string, error) {
	response, err := c.socket.Read()
	if err != nil {
		return "", err
	}

	return response, nil
}

func (c *Client) sendFinishMessage() error {
	if err := c.socket.Write(FINISH_MESSAGE); err != nil {
		return err
	}
	return nil
}

// StartClientLoop Send messages to the client until some time threshold is met
func (c *Client) StartClientLoop() {
	// Create the connection the server in every loop iteration.
	err := c.createClientSocket()
	if err != nil {
		return
	}

	// Go routine to read the data from the file
	go c.readBets()

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

		response, err := c.waitForServerResponse()

		if err != nil {
			log.Criticalf("action: apuesta_enviada | result: fail | cantidad: %v", len(batch))
			break
		}

		if response != SUCCESS_MESSAGE {
			log.Criticalf("action: apuesta_enviada | result: fail | cantidad: %v", len(batch))
			break
		}
	}

	c.sendFinishMessage()
	response, err := c.waitForServerResponse()
	if err != nil {
		log.Criticalf("action: finish | result: fail")
	} else if response != SUCCESS_MESSAGE {
		log.Criticalf("action: finish | result: fail")
	} else {
		log.Infof("action: finish | result: success")
	}

	c.cleanUp()
}

func (c *Client) StopClientLoop() {
	c.running = false
}

func (c *Client) cleanUp() {
	close(c.done)
	close(c.freeBets)
	c.socket.Close()
	log.Infof("action: exit | result: success | client_id: %v", c.config.ID)
}
