package network

import (
	"encoding/binary"

	"github.com/op/go-logging"

	"github.com/7574-sistemas-distribuidos/docker-compose-init/client/repository"
	"github.com/7574-sistemas-distribuidos/docker-compose-init/client/utils"
)

var log = logging.MustGetLogger("log")

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
	MaxBatchLength uint16 = 8000

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

// Converts a slice of bets into a slice of bytes: The payload to be sent to the server.
//
// betsToSend keeps track of the number of bets to send (which may be less that the actual size of the bets slice),
// So the csv pointer can be advanced accurately.
//
// Each bet is converted into its own byte slice (currentBetBytes) and is concatenated
// to the whole batch byte slice (batchBytes) if it doesn't exceed the hard limit of 8kB.
func ConvertBetsToBytes(bets []repository.Bet) ([]byte, uint16, int) {
	var batchBytes []byte
	betsProcessed := 0
	for _, bet := range bets {
		var currentBetBytes []byte
		birthYear, birthMonth, birthDay, err := utils.ParseBirthdate(bet.Birthdate)
		if err != nil {
			// If birthdate parsing fails, log the error and skip this specific bet,
			// whilst incrementing betsProcessed to advance the csv pointer.
			log.Errorf("action: parse_birthdate | result: fail | birthdate: %s | error: %v",
				bet.Birthdate,
				err,
			)
			betsProcessed++
			continue
		}

		firstNameSize := byte(len(bet.FirstName))
		lastNameSize := byte(len(bet.LastName))

		currentBetBytes = append(currentBetBytes, firstNameSize)
		currentBetBytes = append(currentBetBytes, []byte(bet.FirstName)...)

		currentBetBytes = append(currentBetBytes, lastNameSize)
		currentBetBytes = append(currentBetBytes, []byte(bet.LastName)...)

		dniBytes := make([]byte, DNIFieldSize)
		binary.BigEndian.PutUint32(dniBytes, bet.Dni)
		currentBetBytes = append(currentBetBytes, dniBytes...)

		birthYearBytes := make([]byte, BirthYearFieldSize)
		binary.BigEndian.PutUint16(birthYearBytes, birthYear)
		currentBetBytes = append(currentBetBytes, birthYearBytes...)

		currentBetBytes = append(currentBetBytes, birthMonth)
		currentBetBytes = append(currentBetBytes, birthDay)

		betAmountBytes := make([]byte, BetAmountFieldSize)
		binary.BigEndian.PutUint32(betAmountBytes, bet.Number)
		currentBetBytes = append(currentBetBytes, betAmountBytes...)

		// Check if adding the current bet would exceed the max batch length.
		// If so, break the loop and send the current batch.
		if len(batchBytes)+len(currentBetBytes) > int(MaxBatchLength) {
			break
		}
		batchBytes = append(batchBytes, currentBetBytes...)
		betsProcessed++
	}

	return batchBytes, uint16(len(batchBytes)), betsProcessed
}

// Sends a batch of bets to the server through the socket.
//
// First, converts the batch into bytes.
// If batch size is 0, simply returns. Nothing is sent to the server.
// A boolean is returned so the client avoids waiting for a server response, which would block the client.
//
// Sends agency id, then batch size, then batch payload.
//
// Returns a boolean that indicates if a batch was sent or not,
// the number of bets processed (to increment the csv offset accordingly), any errors encountered
func SendBetBatch(socket *Socket, id uint8, bets []repository.Bet) (bool, int, error) {
	batchBytes, batchSize, betsProcessed := ConvertBetsToBytes(bets)

	if batchSize == 0 {
		return false, betsProcessed, nil
	}

	batchSizeBytes := make([]byte, BatchSizeFieldSize)
	binary.BigEndian.PutUint16(batchSizeBytes, batchSize)

	if err := socket.Send_all([]byte{id}); err != nil {
		return false, betsProcessed, err
	}

	if err := socket.Send_all(batchSizeBytes); err != nil {
		return false, betsProcessed, err
	}

	if err := socket.Send_all(batchBytes); err != nil {
		return false, betsProcessed, err
	}

	return true, betsProcessed, nil
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
