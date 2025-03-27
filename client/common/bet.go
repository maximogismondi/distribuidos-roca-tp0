package common

import (
	"fmt"
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

func NewBet(agency string, name string, surname string, document int, birthdate string, number int) *Bet {
	return &Bet{
		Agency:    agency,
		Name:      name,
		Surname:   surname,
		Document:  document,
		Birthdate: birthdate,
		Number:    number,
	}
}

func (b *Bet) String() string {
	params := []string{
		fmt.Sprintf("%v", b.Agency),
		b.Name,
		b.Surname,
		fmt.Sprintf("%v", b.Document),
		b.Birthdate,
		fmt.Sprintf("%v", b.Number),
	}

	return strings.Join(params, BET_SEPARATOR)
}

func FromCSVLine(betString string, agency string) (Bet, error) {
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
		Agency:    agency,
		Name:      params[0],
		Surname:   params[1],
		Document:  document,
		Birthdate: params[3],
		Number:    number,
	}, nil
}
