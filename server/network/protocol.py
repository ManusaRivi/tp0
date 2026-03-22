from enum import IntEnum

from network.socket import Socket

"""

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
    - Bet number: 4 bytes (uint32)

Server response:
- Status: 1 byte (0 for failure, 1 for success)

"""

# =====================
#  Protocol Constants
# =====================

BATCH_SIZE_BYTES = 2

class ClientMessageFieldSize(IntEnum):
    NAME_LENGTH = 1
    ID = 1
    DNI = 4
    BIRTH_YEAR = 2
    BIRTH_MONTH = 1
    BIRTH_DAY = 1
    BET_NUMBER = 4


class ServerMessageStatus(IntEnum):
	FAILURE = 0
	SUCCESS = 1


# =====================
#    Send Bet Result
# =====================


def send_bet_result(client_sock, status: ServerMessageStatus) -> None:
    client_sock.send(status.value.to_bytes(1, byteorder='big'))


# =====================
#  Receive Bet Message
# =====================


def recv_with_prefix(client_sock: Socket, field_name: str) -> str:
    length_bytes = client_sock.recv_all(ClientMessageFieldSize.NAME_LENGTH)
    length = int.from_bytes(length_bytes, byteorder='big')

    raw_text = client_sock.recv_all(length)
    text_without_padding = raw_text.rstrip(b'\x00')

    try:
        return text_without_padding.decode('utf-8')
    except UnicodeDecodeError as error:
        raise ValueError(f'Invalid UTF-8 encoding for {field_name}') from error


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


def parse_batch(id: int, batch_bytes: bytes) -> list[dict]:
    bets = []
    offset = 0
    while offset < len(batch_bytes):
        bet_data = batch_bytes[offset:]
        bet, bytes_read = parse_bet(id, bet_data)
        bets.append(bet)
        offset += bytes_read
    return bets

def receive_bet_batch(client_sock: Socket) -> list[dict]:
    try:
        id = int.from_bytes(client_sock.recv_all(ClientMessageFieldSize.ID), byteorder='big')
        batch_size = int.from_bytes(client_sock.recv_all(BATCH_SIZE_BYTES), byteorder='big')

        batch_bytes = client_sock.recv_all(batch_size)

        return parse_batch(id, batch_bytes)
    except Exception as e:
        raise ValueError(f'Error receiving bet chunk: {e}') from e


def receive_bet(client_sock: Socket):
    id = int.from_bytes(client_sock.recv_all(ClientMessageFieldSize.ID), byteorder='big')
    first_name = recv_with_prefix(client_sock, 'first_name')
    last_name = recv_with_prefix(client_sock, 'last_name')
    dni_number = int.from_bytes(client_sock.recv_all(ClientMessageFieldSize.DNI), byteorder='big')
    birth_year = int.from_bytes(client_sock.recv_all(ClientMessageFieldSize.BIRTH_YEAR), byteorder='big')
    birth_month = int.from_bytes(client_sock.recv_all(ClientMessageFieldSize.BIRTH_MONTH), byteorder='big')
    birth_day = int.from_bytes(client_sock.recv_all(ClientMessageFieldSize.BIRTH_DAY), byteorder='big')
    bet_number = int.from_bytes(client_sock.recv_all(ClientMessageFieldSize.BET_NUMBER), byteorder='big')

    return {
        "id": id, 
        "first_name": first_name,
        "last_name": last_name,
        "dni": dni_number,
        "birthdate": f'{birth_year:04d}-{birth_month:02d}-{birth_day:02d}',
        "number": bet_number
    }