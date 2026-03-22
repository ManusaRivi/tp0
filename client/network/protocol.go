package network

import (
	"encoding/binary"

	"github.com/7574-sistemas-distribuidos/docker-compose-init/client/repository"
	"github.com/7574-sistemas-distribuidos/docker-compose-init/client/utils"
)

/*

Protocol definition:
- Client sends Bet message to Server
- Server responds with a Result (Sucess of Failure)

Data serialization:
- Agency id: 1 byte
- Batch Size: 2 bytes
Batch size is 16 bits because the max batch size is 8kB. 8 bits are not enough, and 24 bits are too much.
- Main payload with bets:
	- First name & Last name: 1 byte for the size of the field + N bytes for the content (max 255 bytes for the content)
	- DNI: 4 bytes (uint32)
	- Birthdate: broken down into three separate fields to minimize bytes sent.
	- Birth year: 2 bytes (uint16)
	- Birth month: 1 byte (uint8)
	- Birth day: 1 byte (uint8)
	- Bet amount: 4 bytes (uint32)

Server response:
- Status: 1 byte (0 for failure, 1 for success)

*/

// =====================
//  Protocol Constants
// =====================

const (
	BatchSizeFieldSize uint8 = 2
	DNIFieldSize       uint8 = 4
	BirthYearFieldSize uint8 = 2
	BetAmountFieldSize uint8 = 4
)

const (
	ServerMessageStatusSize    uint8 = 1
	ServerMessageStatusFailure uint8 = 0
	ServerMessageStatusSuccess uint8 = 1
)

// =====================
//   Send Bet Message
// =====================

func ConvertBetsToBytes(bets []repository.Bet) ([]byte, uint16, error) {
	var batchBytes []byte
	for _, bet := range bets {
		birthYear, birthMonth, birthDay, err := utils.ParseBirthdate(bet.Birthdate)
		if err != nil {
			return nil, 0, err
		}

		firstNameSize := byte(len(bet.FirstName))
		lastNameSize := byte(len(bet.LastName))

		batchBytes = append(batchBytes, firstNameSize)
		batchBytes = append(batchBytes, []byte(bet.FirstName)...)

		batchBytes = append(batchBytes, lastNameSize)
		batchBytes = append(batchBytes, []byte(bet.LastName)...)

		dniBytes := make([]byte, DNIFieldSize)
		binary.BigEndian.PutUint32(dniBytes, bet.Dni)
		batchBytes = append(batchBytes, dniBytes...)

		birthYearBytes := make([]byte, BirthYearFieldSize)
		binary.BigEndian.PutUint16(birthYearBytes, birthYear)
		batchBytes = append(batchBytes, birthYearBytes...)

		batchBytes = append(batchBytes, birthMonth)
		batchBytes = append(batchBytes, birthDay)

		betAmountBytes := make([]byte, BetAmountFieldSize)
		binary.BigEndian.PutUint32(betAmountBytes, bet.Number)
		batchBytes = append(batchBytes, betAmountBytes...)
	}

	return batchBytes, uint16(len(batchBytes)), nil
}

func SendBetBatch(socket *Socket, id uint8, bets []repository.Bet) error {
	batchBytes, batchSize, err := ConvertBetsToBytes(bets)
	if err != nil {
		return err
	}

	batchSizeBytes := make([]byte, BatchSizeFieldSize)
	binary.BigEndian.PutUint16(batchSizeBytes, batchSize)

	if err := socket.Send_all([]byte{id}); err != nil {
		return err
	}

	if err := socket.Send_all(batchSizeBytes); err != nil {
		return err
	}

	if err := socket.Send_all(batchBytes); err != nil {
		return err
	}

	return nil
}

func SendBetMessage(
	socket *Socket,
	id uint8,
	bet repository.Bet,
) error {
	birthYear, birthMonth, birthDay, err := utils.ParseBirthdate(bet.Birthdate)
	if err != nil {
		return err
	}

	if err := socket.Send_all([]byte{id}); err != nil {
		return err
	}

	firstNameSize := byte(len(bet.FirstName))
	lastNameSize := byte(len(bet.LastName))

	if err := socket.Send_all([]byte{firstNameSize}); err != nil {
		return err
	}

	if err := socket.Send_all([]byte(bet.FirstName)); err != nil {
		return err
	}

	if err := socket.Send_all([]byte{lastNameSize}); err != nil {
		return err
	}

	if err := socket.Send_all([]byte(bet.LastName)); err != nil {
		return err
	}

	dniBytes := make([]byte, DNIFieldSize)
	binary.BigEndian.PutUint32(dniBytes, bet.Dni)

	if err := socket.Send_all(dniBytes); err != nil {
		return err
	}

	birthYearBytes := make([]byte, BirthYearFieldSize)
	binary.BigEndian.PutUint16(birthYearBytes, birthYear)

	if err := socket.Send_all(birthYearBytes); err != nil {
		return err
	}

	if err := socket.Send_all([]byte{birthMonth}); err != nil {
		return err
	}

	if err := socket.Send_all([]byte{birthDay}); err != nil {
		return err
	}

	betAmountBytes := make([]byte, BetAmountFieldSize)
	binary.BigEndian.PutUint32(betAmountBytes, bet.Number)

	if err := socket.Send_all(betAmountBytes); err != nil {
		return err
	}

	return nil
}

// =====================
//  Receive Bet Result
// =====================

func ReceiveServerMessage(socket *Socket) (uint8, error) {
	status := make([]byte, ServerMessageStatusSize)
	if err := socket.Receive_all(status); err != nil {
		return 0, err
	}
	return status[0], nil
}
