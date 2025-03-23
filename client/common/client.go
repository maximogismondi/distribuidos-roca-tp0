package common

import (
	"bufio"
	"fmt"
	"net"
	"time"

	"github.com/op/go-logging"
)

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
	message := fmt.Sprintf(
		"[CLIENT %v] %v + %v + %v + %v + %v\n",
		c.config.ID,
		c.config.Name,
		c.config.Surname,
		c.config.Document,
		c.config.Birthdate.Format("2006-01-02"),
		c.config.Number,
	)

	fmt.Fprint(c.conn, message)
	msg, err := bufio.NewReader(c.conn).ReadString('\n')
	c.conn.Close()

	if err != nil {
		log.Errorf("action: receive_message | result: fail | client_id: %v | error: %v",
			c.config.ID,
			err,
		)
		return
	}

	log.Infof("action: receive_message | result: success | client_id: %v | msg: %v",
		c.config.ID,
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
