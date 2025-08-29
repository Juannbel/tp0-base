import socket
import logging
from common.protocol import Protocol
from common.utils import store_bets

NUMBER_OF_AGENCIES = 5

class Server:
    def __init__(self, port, listen_backlog):
        # Initialize server socket
        self._server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self._server_socket.bind(('', port))
        self._server_socket.listen(listen_backlog)
        self._keep_running = True

    def run(self):
        """
        Dummy Server loop

        Server that accept a new connections and establishes a
        communication with a client. After client with communucation
        finishes, servers starts to accept new connections again
        """
        
        stored_bets = 0

        while self._keep_running and stored_bets < NUMBER_OF_AGENCIES:
            client_sock = self.__accept_new_connection()
            if client_sock is not None:
                self.__handle_client_connection(client_sock)
                stored_bets += 1

        logging.info('action: stop_server | result: success')

    def __handle_client_connection(self, client_sock):
        """
        Read bet from a specific client socket and closes the socket

        If a problem arises in the communication with the client, the
        client socket will also be closed
        """
        try:
            protocol = Protocol(client_sock)
            bet = protocol.receive_bet()
            protocol.confirm_reception()

            store_bets([bet])
            logging.info(f'action: apuesta_almacenada | result: success | dni: {bet.document} | numero: {bet.number}')
            
        except OSError as e:
            logging.error(f'action: receive_message | result: fail | error: {e}')
        finally:
            protocol.close()

    def __accept_new_connection(self):
        """
        Accept new connections

        Function blocks until a connection to a client is made.
        Then connection created is printed and returned
        """

        # Connection arrived
        logging.info('action: accept_connections | result: in_progress')
        try:
            c, addr = self._server_socket.accept()
            logging.info(f'action: accept_connections | result: success | ip: {addr[0]}')
            return c
        except OSError as e:
            logging.error(f'action: accept_connections | result: fail | error: {e}')
            return None

    def stop(self, signum, frame):
        """
        Stop the server gracefully
        """
        logging.info('action: stop_server | result: in_progress')
        self._keep_running = False
        self._server_socket.close()
        logging.info('action: close_server_socket | result: success')
        