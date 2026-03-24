package network

import (
	"encoding/binary"
	"fmt"

	"github.com/7574-sistemas-distribuidos/docker-compose-init/client/repository"
	"github.com/7574-sistemas-distribuidos/docker-compose-init/client/utils"
	"github.com/op/go-logging"
)

/*
Payload Sizes per Message Type:

Incoming messages from Client:

- For BET_BATCH: variable, specified in header
- For FINISHED: 1 byte (agency_id)
- For WINNERS_REQUEST: 1 byte (agency_id)

Outgoing messages to Client:
- For ACK: 1 byte (success: 0, failure: 1)
- For WINNERS_RESPONSE: variable, specified in header

Payload Contents per Message Type:
Incoming messages from Client:

- For BET_BATCH: agency ID (1 byte) + a batch of bets of variable size
- For FINISHED: agency ID (1 byte)
- For WINNERS_REQUEST: agency ID (1 byte)

Outgoing messages to Client:
- For ACK: 1 byte (0 for failure, 1 for success)
- For WINNERS_RESPONSE: number of winners (4 bytes) + list of winner DNIs (4 bytes each)
If winners are not yet available, number of winners will be 0, and no DNIs will be sent.
*/

var log = logging.MustGetLogger("log")

const (
	HeaderSize         = 5
	PayloadLengthBytes = 4
	AgencyIDBytes      = 1

	AckStatusBytes    = 1
	WinnersCountBytes = 4
	DNIBytes          = 4
)

type MessageType uint8

const (
	MessageTypeBetBatch       MessageType = 0x01
	MessageTypeAck            MessageType = 0x02
	MessageTypeFinished       MessageType = 0x03
	MessageTypeWinnersRequest MessageType = 0x04
	MessageTypeWinnersResp    MessageType = 0x05
)

const (
	MaxBatchLength uint16 = 8000

	BatchSizeFieldSize uint8 = 2
	DNIFieldSize       uint8 = 4
	BirthYearFieldSize uint8 = 2
	BetAmountFieldSize uint8 = 4
)

type ServerAckStatus uint8

const (
	ServerAckStatusSuccess ServerAckStatus = 0
	ServerAckStatusFailure ServerAckStatus = 1
)

/*
Parses the header of a message received from the Server.
*/
func parseHeader(header []byte) (MessageType, uint32, error) {
	if len(header) != HeaderSize {
		return 0, 0, fmt.Errorf("invalid header length: got %d, expected %d", len(header), HeaderSize)
	}

	msgType := MessageType(header[0])
	payloadLen := binary.BigEndian.Uint32(header[1 : 1+PayloadLengthBytes])

	return msgType, payloadLen, nil
}

/*
Constructs header for a message of type msgType and given payload length.
*/
func buildHeader(msgType MessageType, payloadLen uint32) []byte {
	header := make([]byte, HeaderSize)
	header[0] = byte(msgType)
	binary.BigEndian.PutUint32(header[1:1+PayloadLengthBytes], payloadLen)
	return header
}

/*
Sends generic message of type msgType to the Server.
First, builds header and sends it, then sends payload.
*/
func sendMessage(socket *Socket, msgType MessageType, payload []byte) error {
	header := buildHeader(msgType, uint32(len(payload)))

	if err := socket.Send_all(header); err != nil {
		return err
	}

	if len(payload) == 0 {
		return nil
	}

	return socket.Send_all(payload)
}

/*
Converts a slice of bets into a slice of bytes: The payload to be sent to the server.

betsToSend keeps track of the number of bets to send (which may be less that the actual size of the bets slice),
So the csv pointer can be advanced accurately.

Each bet is converted into its own byte slice (currentBetBytes) and is concatenated
to the whole batch byte slice (batchBytes) if it doesn't exceed the hard limit of 8kB.
*/
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

/*
Builds payload of bet batch. Sends BET_BATCH message to Server.
*/
func SendBetBatch(socket *Socket, agencyID uint8, bets []repository.Bet) (bool, int, error) {
	batchBytes, batchSize, betsProcessed := ConvertBetsToBytes(bets)

	if batchSize == 0 {
		return false, betsProcessed, nil
	}

	payload := make([]byte, 0, AgencyIDBytes+len(batchBytes))
	payload = append(payload, agencyID)
	payload = append(payload, batchBytes...)

	if err := sendMessage(socket, MessageTypeBetBatch, payload); err != nil {
		return false, betsProcessed, err
	}

	return true, betsProcessed, nil
}

/*
Sends FINISHED message to Server.
Payload is simply one byte: the agency ID.
*/
func SendFinishedMessage(socket *Socket, agencyID uint8) error {
	payload := []byte{agencyID}
	return sendMessage(socket, MessageTypeFinished, payload)
}

/*
Sends WINNERS_REQUEST message to Server.
Payload is simply one byte: the agency ID.
*/
func SendWinnersRequestMessage(socket *Socket, agencyID uint8) error {
	payload := []byte{agencyID}
	return sendMessage(socket, MessageTypeWinnersRequest, payload)
}

/*
Receives message header, parses header and receives payload.
Returns:
- Message type
- Payload as raw byte slice
- Any error encountered
*/
func ReceiveMessage(socket *Socket) (MessageType, []byte, error) {
	header := make([]byte, HeaderSize)
	if err := socket.Receive_all(header); err != nil {
		return 0, nil, err
	}

	msgType, payloadLen, err := parseHeader(header)
	if err != nil {
		return 0, nil, err
	}

	payload := make([]byte, int(payloadLen))
	if payloadLen > 0 {
		if err := socket.Receive_all(payload); err != nil {
			return 0, nil, err
		}
	}

	return msgType, payload, nil
}

/*
Receives an ACK message from the server.
Since payload is one byte when message is ACK, there's no need to parse the payload.
*/
func ReceiveAck(socket *Socket) (ServerAckStatus, error) {
	msgType, payload, err := ReceiveMessage(socket)
	if err != nil {
		return 0, err
	}

	if msgType != MessageTypeAck {
		return 0, fmt.Errorf("invalid response type: got %d, expected ACK", msgType)
	}

	if len(payload) != AckStatusBytes {
		return 0, fmt.Errorf("invalid ACK payload length: got %d, expected %d", len(payload), AckStatusBytes)
	}

	status := ServerAckStatus(payload[0])
	return status, nil
}

/*
Parses the payload corresponding to a WINNERS_RESPONSE message.
- winners count may be 0 when the agency has no winners.
- otherwise, count is > 0 and each DNI is 4 bytes.
*/
func ParseWinnersResponsePayload(payload []byte) ([]uint32, error) {
	if len(payload) < WinnersCountBytes {
		return nil, fmt.Errorf("WINNERS_RESPONSE payload too short: got %d", len(payload))
	}

	count := binary.BigEndian.Uint32(payload[:WinnersCountBytes])
	expectedLen := int(WinnersCountBytes) + int(count)*int(DNIBytes)
	if len(payload) != expectedLen {
		return nil, fmt.Errorf(
			"invalid WINNERS_RESPONSE payload length: got %d, expected %d",
			len(payload),
			expectedLen,
		)
	}

	// If count is 0, agency has no winners.
	if count == 0 {
		return []uint32{}, nil
	}

	dnis := make([]uint32, 0, count)
	for offset := int(WinnersCountBytes); offset < len(payload); offset += int(DNIBytes) {
		dni := binary.BigEndian.Uint32(payload[offset : offset+int(DNIBytes)])
		dnis = append(dnis, dni)
	}

	return dnis, nil
}
