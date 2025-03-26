package common

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"github.com/op/go-logging"
)

const BATCH_SEPARATOR = '*'
const WINNERS_SEPARATOR = ','

const FINISH_MESSAGE = "finish"
const SUCCESS_MESSAGE = "success"
const REQUEST_RESULTS_MESSAGE = "request"
const WINNERS_MESSAGE = "winners"
const NOT_READY_MESSAGE = "not_ready"

const MAX_BATCH_BYTES = 8*1024 - 1 // 8kB - 1 (comunicator delimiter)

var log = logging.MustGetLogger("log")

// AgencyConfig Configuration used by the client
type AgencyConfig struct {
	ID            string
	ServerAddress string
	BatchAmount   int
	DataFilePath  string
}

// Agency Entity that encapsulates how
type Agency struct {
	config    AgencyConfig
	socket    Socket
	bets      chan Bet
	freeBets  chan struct{}
	done      chan struct{}
	nextBet   string
	running   bool
	sleepTime time.Duration
}

// NewAgency Initializes a new client receiving the configuration
// as a parameter
func NewAgency(config AgencyConfig, done chan struct{}) *Agency {
	agency := &Agency{
		config:    config,
		bets:      make(chan Bet, config.BatchAmount),
		freeBets:  make(chan struct{}, config.BatchAmount),
		done:      done,
		running:   true,
		sleepTime: 1,
	}

	// Fill the freeBets channel with the amount of bets
	for i := 0; i < config.BatchAmount; i++ {
		agency.freeBets <- struct{}{}
	}

	return agency
}

// CreateClientSocket Initializes client socket. In case of
// failure, error is printed in stdout/stderr and exit 1
// is returned
func (a *Agency) connectToServer() error {
	conn, err := net.Dial("tcp", a.config.ServerAddress)
	if err != nil {
		log.Criticalf(
			"action: connect | result: fail | client_id: %v | error: %v",
			a.config.ID,
			err,
		)
		return err
	}

	socket, err := NewSocket(conn, a.config.ID)
	if err != nil {
		log.Criticalf(
			"action: connect | result: fail | client_id: %v | error: %v",
			a.config.ID,
			err,
		)
		return err
	}

	a.socket = socket

	return nil
}

func (a *Agency) readBets() {
	file, err := os.Open(a.config.DataFilePath)

	if err != nil {
		close(a.bets)
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() && a.running {
		line := scanner.Text()
		bet, err := FromCSVLine(line, a.config.ID)
		if err != nil {
			log.Errorf("Error parsing line: %v", line)
			continue
		}

		select {
		case <-a.freeBets:
			a.bets <- bet
		case <-a.done:
			close(a.bets)
			return
		}
	}
	close(a.bets)
}

func (a *Agency) buildBatch() []string {
	batch := []string{}
	accumlatedBytes := 0

	if a.nextBet != "" {
		batch = append(batch, a.nextBet)
		accumlatedBytes += len(a.nextBet)
		a.nextBet = ""
	}

	for len(batch) < a.config.BatchAmount {
		bet, ok := <-a.bets
		if !ok {
			break
		}

		a.freeBets <- struct{}{}
		betStr := bet.String()
		byteLength := len(betStr)

		if accumlatedBytes+byteLength+1 > MAX_BATCH_BYTES {
			a.nextBet = betStr
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

func (a *Agency) sendBatch(batch []string) error {
	batchStr := strings.Join(batch, string(BATCH_SEPARATOR))

	if err := a.socket.Write(batchStr); err != nil {
		log.Criticalf("action: apuesta_enviada | result: fail | cantidad: %v", len(batch))
		return err
	}

	return nil
}

func (a *Agency) waitForServerResponse() (string, error) {
	response, err := a.socket.Read()
	if err != nil {
		return "", err
	}

	return response, nil
}

func (a *Agency) sendFinishMessage() error {

	if err := a.socket.Write(FINISH_MESSAGE); err != nil {
		return err
	}
	return nil
}

func (a *Agency) sendRequestResultsMessage() error {
	if err := a.socket.Write(REQUEST_RESULTS_MESSAGE); err != nil {
		return err
	}

	return nil
}

func (a *Agency) processResults(results string) (bool, error) {
	fields := strings.Split(results, string(WINNERS_SEPARATOR))

	if len(fields) == 0 {
		return false, fmt.Errorf("invalid number of fields in results")
	}

	if fields[0] == NOT_READY_MESSAGE {
		return false, nil
	}

	if fields[0] != WINNERS_MESSAGE {
		return false, fmt.Errorf("invalid message type in results")
	}

	numberWinners := len(fields) - 1
	log.Infof("action: consulta_ganadores | result: success | cant_ganadores: %v", numberWinners)

	return true, nil
}

func (a *Agency) sendBets() error {
	for a.running {
		batch := a.buildBatch()

		if len(batch) == 0 {
			break
		}

		err := a.sendBatch(batch)
		if err != nil {
			log.Criticalf("action: apuesta_enviada | result: fail | cantidad: %v", len(batch))
			return err
		}

		response, err := a.waitForServerResponse()

		if err != nil {
			log.Criticalf("action: apuesta_enviada | result: fail | cantidad: %v", len(batch))
			return err
		}

		if response != SUCCESS_MESSAGE {
			log.Criticalf("action: apuesta_enviada | result: fail | cantidad: %v", len(batch))
			return fmt.Errorf("server response was not success")
		}
	}

	return nil
}

func (a *Agency) waitForServerResults() error {
	for a.running {
		err := a.sendRequestResultsMessage()

		if err != nil {
			log.Debugf("action: wait_for_results | result: fail | error: %v", err)
			return err
		}

		response, err := a.waitForServerResponse()

		if err != nil {
			log.Debugf("action: wait_for_results2 | result: fail | error: %v", err)
			return err
		}

		done, err := a.processResults(response)

		if err != nil {
			log.Debugf("action: wait_for_results3 | result: fail | error: %v", err)
			return err
		}

		if done {
			break
		}

		// Exponential backoff
		time.Sleep(a.sleepTime * time.Second)
		a.sleepTime *= 2

		err = a.connectToServer()

		if err != nil {
			log.Debugf("action: wait_for_results4 | result: fail | error: %v", err)
			return err
		}
	}

	return nil
}

// StartAgencyLoop Send messages to the client until some time threshold is met
func (a *Agency) StartAgencyLoop() {
	// Create the connection the server in every loop iteration.
	err := a.connectToServer()
	if err != nil {
		log.Criticalf("action: connect | result: fail")
		return
	}

	defer a.cleanUp()

	// Go routine to read the data from the file
	go a.readBets()

	// Build and send batches until the channel is closed (no more bets)
	err = a.sendBets()
	if err != nil {
		log.Criticalf("action: apuesta_enviada | result: fail")
		return
	}

	err = a.sendFinishMessage()
	if err != nil {
		log.Criticalf("action: finish | result: fail")
		return
	}

	// Wait for the server to finish processing the results
	err = a.waitForServerResults()
	if err != nil {
		log.Criticalf("action: wait_for_results | result: fail | error: %v", err)
		return
	}

}

func (a *Agency) StopAgencyLoop() {
	a.running = false
}

func (a *Agency) cleanUp() {
	close(a.done)
	close(a.freeBets)
	a.socket.Close()

	time.Sleep(1 * time.Second)
}
