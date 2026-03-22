package utils

import (
	"time"

	"github.com/op/go-logging"

	"github.com/7574-sistemas-distribuidos/docker-compose-init/client/repository"
)

var log = logging.MustGetLogger("log")

func ParseBirthdate(birthdate string) (uint16, uint8, uint8, error) {
	layout := "2006-01-02"
	t, err := time.Parse(layout, birthdate)
	if err != nil {
		return 0, 0, 0, err
	}

	return uint16(t.Year()), uint8(t.Month()), uint8(t.Day()), nil
}

func PrintBet(bet repository.Bet) {
	log.Debugf("action: bet_data | result: success | first_name: %s, last_name: %s, dni: %v, birthdate: %s, number: %v",
		bet.FirstName,
		bet.LastName,
		bet.Dni,
		bet.Birthdate,
		bet.Number,
	)
}
