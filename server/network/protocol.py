from enum import IntEnum

from server.network.socket import Socket

# =====================
#  Protocol Constants
# =====================

class ClientMessageFieldSize(IntEnum):
	FIRST_NAME = 1
	LAST_NAME = 1
	DNI = 4
	BIRTH_YEAR = 2
	BIRTH_MONTH = 1
	BIRTH_DAY = 1
	BET_AMOUNT = 4


class ServerMessageStatus(IntEnum):
	FAILURE = 0
	SUCCESS = 1


# =====================
#    Send Bet Result
# =====================


def send_bet_result(client_sock, status: ServerMessageStatus):
    client_sock.send(status.value.to_bytes(1, byteorder='big'))


# =====================
#  Receive Bet Message
# =====================


def recv_with_prefix(client_sock: Socket, field_name: str) -> str:
    length_bytes = client_sock.recv_all(1)
    length = int.from_bytes(length_bytes, byteorder='big')

    raw_text = client_sock.recv_all(length)
    text_without_padding = raw_text.rstrip(b'\x00')

    try:
        return text_without_padding.decode('utf-8')
    except UnicodeDecodeError as error:
        raise ValueError(f'Invalid UTF-8 encoding for {field_name}') from error
	

def receive_bet(client_sock: Socket):
    first_name = recv_with_prefix(client_sock, 'first_name')
    last_name = recv_with_prefix(client_sock, 'last_name')
    dni_number = int.from_bytes(
        client_sock.recv_all(ClientMessageFieldSize.DNI),
        byteorder='big',
    )
    birth_year = int.from_bytes(
        client_sock.recv_all(ClientMessageFieldSize.BIRTH_YEAR),
        byteorder='big',
    )
    birth_month = int.from_bytes(
        client_sock.recv_all(ClientMessageFieldSize.BIRTH_MONTH),
        byteorder='big',
    )
    birth_day = int.from_bytes(
        client_sock.recv_all(ClientMessageFieldSize.BIRTH_DAY),
        byteorder='big',
    )
    bet_amount = int.from_bytes(
        client_sock.recv_all(ClientMessageFieldSize.BET_AMOUNT),
        byteorder='big',
    )

    return {
        "first_name": first_name,
        "last_name": last_name,
        "dni": dni_number,
        "birthdate": f'{birth_year:04d}-{birth_month:02d}-{birth_day:02d}',
        "bet_amount": bet_amount
    }