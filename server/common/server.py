import socket
import logging
from network import protocol
from network.socket import Socket
from common import utils

class Server:
    def __init__(self, port, listen_backlog):
        # Initialize server socket
        self._server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self._server_socket.bind(('', port))
        self._server_socket.listen(listen_backlog)
        self.stopped = False

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

    def __handle_client_connection(self, client_sock: Socket):
        """
        Read message from a specific client socket and closes the socket

        If a problem arises in the communication with the client, the
        client socket will also be closed
        """

        try:
            bet = protocol.receive_bet(client_sock)
            utils.store_bets([utils.Bet(
                agency=0,  # TODO: Get agency from client
                first_name=bet['first_name'],
                last_name=bet['last_name'],
                document=str(bet['dni']),
                birthdate=bet['birthdate'],
                number=bet['bet_amount'],
            )])
            logging.info(f"action: apuesta_almacenada | result: success | dni: {bet['dni']} | numero: {bet['bet_amount']}")
            protocol.send_bet_result(client_sock, protocol.ServerMessageStatus.SUCCESS)
        except OSError as e:
            logging.error("action: apuesta_almacenada | result: fail | error: {e}")
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
