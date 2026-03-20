package network

import "encoding/binary"

/*

Protocol definition:
- Client sends Bet message to Server
- Server responds with a Result (Sucess of Failure)

Data serialization:
- Agency id: 1 byte
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

func SendBetMessage(
	socket *Socket,
	id uint8,
	firstName string,
	lastName string,
	dni uint32,
	birthYear uint16,
	birthMonth uint8,
	birthDay uint8,
	betAmount uint32,
) error {
	if err := socket.Send_all([]byte{id}); err != nil {
		return err
	}

	firstNameSize := byte(len(firstName))
	lastNameSize := byte(len(lastName))

	if err := socket.Send_all([]byte{firstNameSize}); err != nil {
		return err
	}

	if err := socket.Send_all([]byte(firstName)); err != nil {
		return err
	}

	if err := socket.Send_all([]byte{lastNameSize}); err != nil {
		return err
	}

	if err := socket.Send_all([]byte(lastName)); err != nil {
		return err
	}

	dniBytes := make([]byte, DNIFieldSize)
	binary.BigEndian.PutUint32(dniBytes, dni)

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
	binary.BigEndian.PutUint32(betAmountBytes, betAmount)

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
