package common

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

const BET_SEPARATOR = "+"
const CSV_SEPARATOR = ","
const NUMBER_OF_FIELDS = 5

type Bet struct {
	Agency    string
	Name      string
	Surname   string
	Document  int
	Birthdate string
	Number    int
}

func NewBet(name string, surname string, document int, birthdate string, number int) Bet {
	return Bet{
		Name:      name,
		Surname:   surname,
		Document:  document,
		Birthdate: birthdate,
		Number:    number,
	}
}

func (b *Bet) Encode() string {
	params := []string{
		b.Name,
		b.Surname,
		fmt.Sprintf("%v", b.Document),
		b.Birthdate,
		fmt.Sprintf("%v", b.Number),
	}

	return strings.Join(params, BET_SEPARATOR)
}

func fromCSVLine(betString string) (Bet, error) {
	params := strings.Split(betString, CSV_SEPARATOR)

	if len(params) != NUMBER_OF_FIELDS {
		return Bet{}, fmt.Errorf("invalid number of fields in bet string")
	}

	document, err := strconv.Atoi(params[2])
	if err != nil {
		return Bet{}, fmt.Errorf("invalid document field")
	}

	number, err := strconv.Atoi(params[4])
	if err != nil {
		return Bet{}, fmt.Errorf("invalid number field")
	}

	return Bet{
		Name:      params[0],
		Surname:   params[1],
		Document:  document,
		Birthdate: params[3],
		Number:    number,
	}, nil
}

func ReadBetsFromFile(
	filePath string,
	bets chan Bet,
	freeBets chan struct{},
	done chan struct{},
) {
	file, err := os.Open(filePath)
	if err != nil {
		close(bets)
		return
	}
	defer file.Close()
	defer close(bets)

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		bet, err := fromCSVLine(line)

		if err != nil {
			log.Errorf("Error parsing line: %v", line)
			continue
		}

		// Wait until there is a free bet slot or the agency is stopped
		select {
		case <-freeBets:
			bets <- bet
		case <-done:
			return
		}
	}
}
