package common

import (
	"fmt"
	"strconv"
	"time"

	"github.com/spf13/viper"
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
	log.Debugf("action: agency_data | result: success | first_name: %s, last_name: %s, dni: %v, birth_year: %v, birth_month: %v, birth_day: %v, number: %v",
		agencyData.FirstName,
		agencyData.LastName,
		agencyData.DNI,
		agencyData.BirthYear,
		agencyData.BirthMonth,
		agencyData.BirthDay,
		agencyData.Number,
	)
}

func GetAgencyData(v *viper.Viper) (AgencyData, error) {
	firstName := v.GetString("agency.first_name")
	lastName := v.GetString("agency.last_name")
	dni := v.GetString("agency.dni")

	dniInt, err := strconv.ParseUint(dni, 10, 32)
	if err != nil {
		return AgencyData{}, fmt.Errorf("invalid DNI value %q: %w", dni, err)
	}

	dniUint32 := uint32(dniInt)

	birthdate := v.GetString("agency.birthdate")

	year, month, day, err := parseBirthdate(birthdate)
	if err != nil {
		return AgencyData{}, fmt.Errorf("invalid NACIMIENTO value %q (expected YYYY-MM-DD): %w", birthdate, err)
	}

	number := v.GetString("agency.number")

	numberInt, err := strconv.ParseUint(number, 10, 32)
	if err != nil {
		return AgencyData{}, fmt.Errorf("invalid NUMERO value %q: %w", number, err)
	}

	numberUint32 := uint32(numberInt)

	return AgencyData{
		FirstName:  firstName,
		LastName:   lastName,
		DNI:        dniUint32,
		BirthYear:  year,
		BirthMonth: month,
		BirthDay:   day,
		Number:     numberUint32,
	}, nil
}
