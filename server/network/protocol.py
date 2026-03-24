from enum import IntEnum

from network.socket import Socket

"""
Header:
- Message Type: 1 byte
- Payload Length: 4 bytes (uint32, big endian)
"""
HEADER_SIZE = 5
PAYLOAD_LENGTH_BYTES = 4
AGENCY_ID_BYTES = 1
ACK_STATUS_BYTES = 1
WINNERS_COUNT_BYTES = 4
DNI_BYTES = 4

class MessageType(IntEnum):
    BET_BATCH = 0x01
    ACK = 0x02
    FINISHED = 0x03
    WINNERS_REQUEST = 0x04
    WINNERS_RESPONSE = 0x05

class ServerAckStatus(IntEnum):
	SUCCESS = 0
	FAILURE = 1
"""
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
"""

class ClientMessageFieldSize(IntEnum):
    NAME_LENGTH = 1
    ID = 1
    DNI = 4
    BIRTH_YEAR = 2
    BIRTH_MONTH = 1
    BIRTH_DAY = 1
    BET_NUMBER = 4

"""
Parses a received header's data.
Returns:
- Message Type (so calling function can dispatch the right parsing function)
- Agency ID (to identify the client)
- Payload Length (to know how many bytes to read for the payload)
"""
def parse_header(header_bytes: bytes) -> tuple[MessageType, int]:
    msg_type = MessageType(header_bytes[0])
    payload_len = int.from_bytes(header_bytes[1:5], byteorder='big')

    return msg_type, payload_len

"""
Parses a single bet (array of bytes) and returns:
- A tuple of the bet as a dict
- The number of bytes read from the input.
"""
def parse_bet(id: int, bet_data: bytes) -> tuple[dict, int]:
    offset = 0

    first_name_length = int.from_bytes(bet_data[offset:offset + ClientMessageFieldSize.NAME_LENGTH], byteorder='big')
    offset += ClientMessageFieldSize.NAME_LENGTH
    first_name = bet_data[offset:offset + first_name_length].decode('utf-8')
    offset += first_name_length

    last_name_length = int.from_bytes(bet_data[offset:offset + ClientMessageFieldSize.NAME_LENGTH], byteorder='big')
    offset += ClientMessageFieldSize.NAME_LENGTH
    last_name = bet_data[offset:offset + last_name_length].decode('utf-8')
    offset += last_name_length

    dni_number = int.from_bytes(bet_data[offset:offset + ClientMessageFieldSize.DNI], byteorder='big')
    offset += ClientMessageFieldSize.DNI

    birth_year = int.from_bytes(bet_data[offset:offset + ClientMessageFieldSize.BIRTH_YEAR], byteorder='big')
    offset += ClientMessageFieldSize.BIRTH_YEAR

    birth_month = int.from_bytes(bet_data[offset:offset + ClientMessageFieldSize.BIRTH_MONTH], byteorder='big')
    offset += ClientMessageFieldSize.BIRTH_MONTH

    birth_day = int.from_bytes(bet_data[offset:offset + ClientMessageFieldSize.BIRTH_DAY], byteorder='big')
    offset += ClientMessageFieldSize.BIRTH_DAY

    bet_number = int.from_bytes(bet_data[offset:offset + ClientMessageFieldSize.BET_NUMBER], byteorder='big')
    offset += ClientMessageFieldSize.BET_NUMBER

    return {
        "id": id, 
        "first_name": first_name,
        "last_name": last_name,
        "dni": dni_number,
        "birthdate": f'{birth_year:04d}-{birth_month:02d}-{birth_day:02d}',
        "number": bet_number
    }, offset

"""
Parses a batch of bets (array of bytes) and returns a list of bets as dicts.
"""
def parse_batch(id: int, batch_bytes: bytes) -> list[dict]:
    bets = []
    offset = 0
    while offset < len(batch_bytes):
        bet_data = batch_bytes[offset:]
        bet, bytes_read = parse_bet(id, bet_data)
        bets.append(bet)
        offset += bytes_read
    return bets

"""
Receives a header, and then the payload bytes.
Returns:
- Message Type (so calling function can dispatch the right parsing function)
- Agency ID (to identify the client)
- Raw payload bytes
"""
def receive_message(client_sock: Socket) -> tuple[MessageType, bytes]:
    try:
        header_bytes = client_sock.recv_all(HEADER_SIZE)

        msg_type, payload_len = parse_header(header_bytes)

        # Receive agency id, only 1 byte -> taking first element of array works here
        agency_id = client_sock.recv_all(AGENCY_ID_BYTES)[0]

        # Receive the rest of the payload -> only 1 byte less that total payload length
        payload_bytes = client_sock.recv_all(payload_len - AGENCY_ID_BYTES)

        return msg_type, agency_id, payload_bytes
    except Exception as e:
        raise ValueError(f'Error receiving message: {e}') from e

"""
Sends ACK message to Client.
- Status: 0 for success, 1 for failure
- Payload Length: 4 bytes, of value 1 (payload is only one byte: the status)
"""
def send_ack_message(client_sock: Socket, status: ServerAckStatus) -> None:
    header = bytes([MessageType.ACK]) + ACK_STATUS_BYTES.to_bytes(PAYLOAD_LENGTH_BYTES, byteorder='big')
    client_sock.send_all(header)
    client_sock.send_all(status.value.to_bytes(ACK_STATUS_BYTES, byteorder='big'))

"""
Sends WINNERS_RESPONSE message to Client.
- Payload Length: 4 bytes (number of winners) + 4 bytes per winner (DNI)
- Payload Contents: number of winners (4 bytes) + list of winner DNIs (4 bytes each)
"""
def send_winners_response(client_sock: Socket, dnis: list[int]) -> None:
    payload = len(dnis).to_bytes(WINNERS_COUNT_BYTES, byteorder='big')
    for dni in dnis:
        payload += dni.to_bytes(DNI_BYTES, byteorder='big')

    header = bytes([MessageType.WINNERS_RESPONSE]) + len(payload).to_bytes(PAYLOAD_LENGTH_BYTES, byteorder='big')
    client_sock.send_all(header)
    client_sock.send_all(payload)
