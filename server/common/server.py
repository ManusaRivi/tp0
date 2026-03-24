import socket
import logging
from network import protocol
from network.socket import Socket
from common import utils

MAX_AGENCIES = 5

class Server:
    def __init__(self, port, listen_backlog):
        # Initialize server socket
        self._server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self._server_socket.bind(('', port))
        self._server_socket.listen(listen_backlog)
        self.stopped = False
        self.finished_agencies = set()
        self.winners_by_agency: dict[list] = {}

    def run(self):
        """
        Dummy Server loop

        Server that accept a new connections and establishes a
        communication with a client. After client with communucation
        finishes, servers starts to accept new connections again
        """

        while not self.stopped:
            client_sock = self.__accept_new_connection()
            if not self.stopped or client_sock is not None:
                self.__handle_client_connection(client_sock)
    
    def stop(self):
        """
        Break server loop, close socket so __accept_new_connection is unblocked
        """

        self.stopped = True
        self._server_socket.close()
        logging.info("action: shutdown_server | result: success")

    def __process_winners(self):
        for bet in utils.load_bets():
            if utils.has_won(bet):
                if not bet.agency in self.winners_by_agency:
                    self.winners_by_agency[bet.agency] = []
                self.winners_by_agency[bet.agency].append(bet.document)

    def __handle_client_connection(self, client_sock: Socket):
        """
        Read message from a specific client socket and closes the socket

        If a problem arises in the communication with the client, the
        client socket will also be closed
        """

        try:
            msg_type, agency_id, payload_bytes = protocol.receive_message(client_sock)
            if msg_type == protocol.MessageType.BET_BATCH:
                bets = protocol.parse_batch(agency_id, payload_bytes)
                utils.store_bets([utils.Bet(
                    agency=bet['id'],
                    first_name=bet['first_name'],
                    last_name=bet['last_name'],
                    document=str(bet['dni']),
                    birthdate=bet['birthdate'],
                    number=str(bet['number']))
                    for bet in bets
                ])
                logging.info(f"action: apuesta_recibida | result: success | cantidad: {len(bets)}")
                protocol.send_ack_message(client_sock, protocol.ServerAckStatus.SUCCESS)
            elif msg_type == protocol.MessageType.FINISHED:
                # Mark agency as finished, so when all agencies are finished, we can calculate winners
                logging.info(f"action: envio_finalizado | result: success | agencia: {agency_id}")
                self.finished_agencies.add(agency_id)
                protocol.send_ack_message(client_sock, protocol.ServerAckStatus.SUCCESS)
                # Check if all agencies finished sending bets. If so, find winners and assign them to their corresponding agencies.
                if self.finished_agencies == set(range(1, MAX_AGENCIES + 1)):
                    logging.info("action: todos_envios_finalizados | result: success")
                    self.__process_winners()

            elif msg_type ==  protocol.MessageType.WINNERS_REQUEST:
                # If not all agencies finished sending bets, we cannot calculate winners, so we return empty list.
                # TODO: add new ACK status for "In progress"
                if len(self.finished_agencies) < MAX_AGENCIES:
                    logging.info(f"action: consulta_ganadores | result: fail | agencia: {agency_id}")
                    protocol.send_winners_response(client_sock, [])
                    return
                # Check if all agencies finished sending bets.
                # Fetch winners for that agency based on agency_id
                winners = self.winners_by_agency.get(agency_id, [])
                logging.info(f"action: consulta_ganadores | result: success | agencia: {agency_id}")
                protocol.send_winners_response(client_sock, winners)
            else:
                logging.error(f"action: mensaje_desconocido | result: fail | tipo: {msg_type}")
                protocol.send_ack_message(client_sock, protocol.ServerAckStatus.FAILURE)

        except OSError as e:
            logging.error(f"action: apuesta_recibida | result: fail | cantidad: {len(bets)}")
            protocol.send_ack_message(client_sock, protocol.ServerAckStatus.FAILURE)
        finally:
            client_sock.close()

    def __accept_new_connection(self):
        """
        Accept new connections

        Function blocks until a connection to a client is made.
        Then connection created is printed and returned
        """

        try:
            # Connection arrived
            logging.info('action: accept_connections | result: in_progress')
            c, addr = self._server_socket.accept()
            logging.info(f'action: accept_connections | result: success | ip: {addr[0]}')
            return Socket(c)
        except OSError as e:
            # Bad file descriptor. This means this socket was closed to stop the server
            if e.errno == 9:
                logging.info("action: end_connections | result: success")
            else:
                logging.info(f"action: accept_connections | result: fail | error: [Errno {e.errno}] {e.strerror}")
