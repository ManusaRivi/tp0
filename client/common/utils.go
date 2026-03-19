package common

import (
	"os"
	"strconv"
	"time"
)

func parseBirthdate(birthdate string) (uint16, uint8, uint8, error) {
	layout := "2006-01-02"
	t, err := time.Parse(layout, birthdate)
	if err != nil {
		return 0, 0, 0, err
	}

	return uint16(t.Year()), uint8(t.Month()), uint8(t.Day()), nil
}

func PrintAgencyData(agencyData AgencyData) {
	log.Infof("action: agency_data | first_name: %s, last_name: %s, dni: %v, birth_year: %v, birth_month: %v, birth_day: %v, number: %v",
		agencyData.FirstName,
		agencyData.LastName,
		agencyData.DNI,
		agencyData.BirthYear,
		agencyData.BirthMonth,
		agencyData.BirthDay,
		agencyData.Number,
	)
}

func GetAgencyData() (AgencyData, error) {
	first_name := os.Getenv("NOMBRE")

	last_name := os.Getenv("APELLIDO")

	dni := os.Getenv("DNI")
	dni_int, err := strconv.ParseUint(dni, 10, 32)
	if err != nil {
		return AgencyData{}, err
	}

	dni_int32 := uint32(dni_int)

	birthdate := os.Getenv("NACIMIENTO")

	year, month, day, err := parseBirthdate(birthdate)
	if err != nil {
		log.Criticalf("action: parse_birthdate | result: fail | error: %v",
			err,
		)
		return AgencyData{}, err
	}

	birth_year := uint16(year)
	birth_month := uint8(month)
	birth_day := uint8(day)

	number := os.Getenv("NUMERO")
	number_int, err := strconv.ParseUint(number, 10, 32)
	if err != nil {
		log.Criticalf("action: parse_number | result: fail | error: %v",
			err,
		)
		return AgencyData{}, err
	}

	number_int32 := uint32(number_int)

	return AgencyData{
		FirstName:  first_name,
		LastName:   last_name,
		DNI:        dni_int32,
		BirthYear:  birth_year,
		BirthMonth: birth_month,
		BirthDay:   birth_day,
		Number:     number_int32,
	}, nil
}
