package repository

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
)

const (
	BetFieldsCount = 5

	FirstName = 0
	LastName  = 1
	DNI       = 2
	Birthdate = 3
	Number    = 4
)

type Repository struct {
	betsPath string
}

type Bet struct {
	FirstName string
	LastName  string
	Dni       uint32
	Birthdate string
	Number    uint32
}

func NewRepository(bets_path string) *Repository {
	return &Repository{
		betsPath: bets_path,
	}
}

func (r *Repository) FetchBets(batchSize int) ([]Bet, error) {
	// For now, fetch batchSize bets from csv file.
	// TODO: add a cursor to keep track of the last read position,
	// and fetch the next batch of bets on each call.
	betsCsv, err := os.Open(r.betsPath)
	if err != nil {
		return nil, err
	}
	defer betsCsv.Close()

	csvReader := csv.NewReader(betsCsv)
	csvReader.FieldsPerRecord = BetFieldsCount

	bets := make([]Bet, 0, batchSize)

	for i := 0; i < batchSize; i++ {
		record, err := csvReader.Read()
		if err != nil {
			fmt.Printf("error reading csv: %v\n", err)
			return nil, err
		}

		fmt.Printf("record: %q\n", record)

		dniUint, err := strconv.ParseUint(record[DNI], 10, 32)
		if err != nil {
			return nil, err
		}

		dni32 := uint32(dniUint)

		numberUint, err := strconv.ParseUint(record[Number], 10, 32)
		if err != nil {
			return nil, err
		}

		number32 := uint32(numberUint)

		bet := Bet{
			FirstName: record[FirstName],
			LastName:  record[LastName],
			Dni:       dni32,
			Birthdate: record[Birthdate],
			Number:    number32,
		}
		bets = append(bets, bet)
	}

	return bets, nil
}
