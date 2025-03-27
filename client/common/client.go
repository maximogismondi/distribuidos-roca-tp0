package common

import (
	"fmt"
	"time"

	"github.com/7574-sistemas-distribuidos/docker-compose-init/client/communication"
	"github.com/op/go-logging"
)

const DELAY_MULTIPLIER = 2
const MAX_DELAY = 64 // max retries = 6

var log = logging.MustGetLogger("log")

type AgencyConfig struct {
	ID            string
	ServerAddress string
	BatchAmount   int
	DataFilePath  string
}

type Agency struct {
	config        AgencyConfig
	bets          chan Bet
	freeBets      chan struct{}
	done          chan struct{}
	nextBet       string
	running       bool
	sleepTime     time.Duration
	server_socket *communication.ServerSocket
}

func NewAgency(config AgencyConfig, done chan struct{}) Agency {
	freeBets := make(chan struct{}, config.BatchAmount)

	// Fill the freeBets channel with the amount of bets
	for i := 0; i < config.BatchAmount; i++ {
		freeBets <- struct{}{}
	}

	return Agency{
		config:    config,
		bets:      make(chan Bet, config.BatchAmount),
		freeBets:  freeBets,
		done:      done,
		running:   true,
		sleepTime: 1,
	}
}

// Run Send messages to the client until some time threshold is met
func (a *Agency) Run() {
	// Create the connection the server in every loop iteration.
	err := a.connectToServer()
	if err != nil {
		log.Criticalf("action: connect | result: fail")
		return
	}

	// Close the connection when the function ends
	defer a.cleanUp()

	// Go routine to read the data from the file
	go ReadBetsFromFile(a.config.DataFilePath, a.config.ID, a.bets, a.freeBets, a.done)

	// Build and send batches until the channel is closed (no more bets)
	err = a.sendBets()
	if err != nil {
		return
	}

	// Send the finish message to the server
	err = a.sendFinishMessage()
	if err != nil {
		log.Criticalf("action: finish | result: fail")
		return
	}

	// Wait for the server to finish processing the results
	err = a.fetchServerResults()
	if err != nil {
		log.Criticalf("action: wait_for_results | result: fail | error: %v", err)
		return
	}

}

func (a *Agency) Stop() {
	a.running = false
}

func (a *Agency) connectToServer() error {
	server_socket, err := communication.Connect(a.config.ServerAddress, a.config.ID)
	if err != nil {
		return err
	}

	a.server_socket = &server_socket
	return nil
}

func (a *Agency) disconnectFromServer() {
	if a.server_socket != nil {
		a.server_socket.Close()
	}

	a.server_socket = nil
}

func (a *Agency) waitForServerResponse() (string, string, error) {
	msg, err := a.server_socket.Read()

	if err != nil {
		return "", "", err
	}

	return communication.DecodeMessage(msg)
}

func (a *Agency) waitForSuccessServerResponse() error {
	header, _, err := a.waitForServerResponse()
	if err != nil {
		return err
	}

	if header != communication.SUCCESS_MESSAGE {
		return fmt.Errorf("server response was not success")
	}

	return nil
}

func (a *Agency) waitForResultServerResponse() (bool, string, error) {
	header, payload, err := a.waitForServerResponse()
	if err != nil {
		return false, "", err
	}

	if header == communication.NOT_READY_MESSAGE {
		return false, "", nil
	}

	if header == communication.WINNERS_MESSAGE {
		return true, payload, nil
	}

	return false, "", fmt.Errorf("invalid message type in results")
}

func (a *Agency) sendEncodedMessage(message string) error {
	if a.server_socket == nil {
		return fmt.Errorf("server socket is nil")
	}

	return a.server_socket.Write(message)
}

func (a *Agency) sendFinishMessage() error {
	msg := communication.EncodedFinishMessage()
	return a.sendEncodedMessage(msg)
}

func (a *Agency) sendRequestResultsMessage() error {
	msg := communication.EncodedRequestResultsMessage()
	return a.sendEncodedMessage(msg)
}

func (a *Agency) sendBets() error {
	for a.running {
		batch := a.buildBatch()

		if len(batch) == 0 {
			break
		}

		msg := communication.EncodedBetBatchMessage(batch)
		err := a.server_socket.Write(msg)
		if err != nil {
			log.Criticalf("action: apuesta_enviada | result: fail | cantidad: %v", len(batch))
			return err
		}

		err = a.waitForSuccessServerResponse()

		if err != nil {
			log.Criticalf("action: apuesta_enviada | result: fail | cantidad: %v", len(batch))
			return err
		}
	}

	return nil
}

func (a *Agency) buildBatch() []string {
	batch := []string{}
	accumlatedBytes := 0
	for len(batch) < a.config.BatchAmount {
		if a.nextBet != "" {
			batch = append(batch, a.nextBet)
			accumlatedBytes += len(a.nextBet)
			a.nextBet = ""
		}

		if len(batch) > 1 {
			accumlatedBytes++ // For the batch separator
		}

		bet, ok := <-a.bets
		if !ok {
			break
		}

		a.freeBets <- struct{}{}
		a.nextBet = bet.Encode()

		if !communication.CanAppendBetToBatch(accumlatedBytes, a.nextBet) {
			break
		}

	}

	return batch
}

func (a *Agency) fetchServerResults() error {
	for a.running {
		if a.sleepTime > MAX_DELAY {
			return fmt.Errorf("timeout")
		}

		if a.server_socket == nil {
			if err := a.connectToServer(); err != nil {
				return err
			}
		}

		err := a.sendRequestResultsMessage()
		if err != nil {
			return err
		}

		done, response, err := a.waitForResultServerResponse()

		if err != nil {
			return err
		}

		if !done {
			a.disconnectFromServer()

			// Exponential backoff
			a.sleepTime *= DELAY_MULTIPLIER
			time.Sleep(a.sleepTime * time.Second)
			continue
		}

		a.processWinnersMessage(response)
		break
	}

	return nil
}

func (a *Agency) processWinnersMessage(message string) {
	winners := communication.DecodeWinnersMessage(message)

	numberWinners := len(winners)
	if winners[0] == "" {
		numberWinners = 0
	}

	log.Infof("action: consulta_ganadores | result: success | cant_ganadores: %v", numberWinners)
}

func (a *Agency) cleanUp() {
	close(a.done)
	close(a.freeBets)
	a.disconnectFromServer()

	// For testing purposes
	time.Sleep(1 * time.Second)
}
