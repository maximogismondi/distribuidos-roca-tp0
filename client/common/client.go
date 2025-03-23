package common

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/op/go-logging"
)

const DELIMITER = "+"
const SUCCESS_MESSAGE = "success"

var log = logging.MustGetLogger("log")

// ClientConfig Configuration used by the client
type ClientConfig struct {
	ID            string
	ServerAddress string
	Name          string
	Surname       string
	Document      int
	Birthdate     time.Time
	Number        int
}

// BetInfo Information about the bet
type BetInfo struct {
}

// Client Entity that encapsulates how
type Client struct {
	config  ClientConfig
	conn    net.Conn
	running bool
}

// NewClient Initializes a new client receiving the configuration
// as a parameter
func NewClient(config ClientConfig) *Client {
	client := &Client{
		config:  config,
		running: true,
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
	c.conn = conn
	return nil
}

// StartClientLoop Send messages to the client until some time threshold is met
func (c *Client) StartClientLoop() {
	// Create the connection the server in every loop iteration.
	c.createClientSocket()

	// TODO: Modify the send to avoid short-write

	message_params := []string{
		"AGENCY",
		c.config.ID,
		c.config.Name,
		c.config.Surname,
		fmt.Sprintf("%v", c.config.Document),
		c.config.Birthdate.Format("2006-01-02"),
		fmt.Sprintf("%v", c.config.Number),
	}

	message := strings.Join(message_params, DELIMITER) + "\n"
	fmt.Fprint(c.conn, message)

	msg, err := bufio.NewReader(c.conn).ReadString('\n')
	c.conn.Close()

	if err != nil || msg != SUCCESS_MESSAGE+"\n" {
		log.Errorf("action: apuesta_enviada | result: fail | dni: %v | error: %v",
			c.config.Document,
			err,
		)
		return
	}

	log.Infof("action: apuesta_enviada | result: success | dni: %v | numero: %v",
		c.config.Document,
		msg,
	)

	c.cleanUp()
}

func (c *Client) StopClientLoop() {
	c.running = false
}

func (c *Client) cleanUp() {
	c.conn.Close()
	log.Infof("action: exit | result: success | client_id: %v", c.config.ID)
}
