package repository

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
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
	offset   int
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
		offset:   0,
	}
}

// FetchBets creates a csv reader, advances it to the current offset,
// and reads a batch of bets. It returns the batch, a boolean indicating
// if the end of the file was reached, and any error encountered during the process.
func (r *Repository) FetchBets(batchSize int) ([]Bet, bool, error) {
	betsCsv, err := os.Open(r.betsPath)
	if err != nil {
		return nil, false, err
	}
	defer betsCsv.Close()

	csvReader := csv.NewReader(betsCsv)
	csvReader.FieldsPerRecord = BetFieldsCount

	for i := 0; i < r.offset; i++ {
		_, err := csvReader.Read()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return []Bet{}, true, nil
			}
			return nil, false, err
		}
	}

	bets := make([]Bet, 0, batchSize)

	for i := 0; i < batchSize; i++ {
		record, err := csvReader.Read()

		if err != nil {
			if errors.Is(err, io.EOF) {
				return bets, true, nil
			}
			fmt.Printf("error reading csv: %v\n", err)
			return nil, false, err
		}

		dniUint, err := strconv.ParseUint(record[DNI], 10, 32)
		if err != nil {
			return nil, false, err
		}

		dni32 := uint32(dniUint)

		numberUint, err := strconv.ParseUint(record[Number], 10, 32)
		if err != nil {
			return nil, false, err
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

	return bets, false, nil
}

// Advances the offset by the amount of bets processed.
func (r *Repository) AdvanceBatch(betsProcessed int) {
	r.offset += betsProcessed
}
