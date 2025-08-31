import socket
import logging
from common.protocol import Protocol
from common.utils import store_bets

class Server:
    def __init__(self, port, listen_backlog, number_of_agencies):
        # Initialize server socket
        self._server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self._server_socket.bind(('', port))
        self._server_socket.listen(listen_backlog)
        self._keep_running = True
        self._number_of_agencies = number_of_agencies
        self._processed_agencies = set()
        
    def run(self):
        """
        Dummy Server loop

        Server that accept a new connections and establishes a
        communication with a client. After client with communucation
        finishes, servers starts to accept new connections again
        """
        
        while self._keep_running:
            client_sock = self.__accept_new_connection()
            if client_sock is not None:
                self.__handle_client_connection(client_sock)

        logging.info('action: stop_server | result: success')

    def __handle_client_connection(self, client_sock):
        try:
            protocol = Protocol(client_sock)
            
            logging.debug('action: receive_bets_batches | result: in_progress')

            while self._keep_running:
                bets_batch = protocol.receive_bets_batch()
                
                if not bets_batch:
                    logging.debug('action: receive_bets_batches | result: success | info: no more bets')
                    break
            
                protocol.confirm_reception()

                store_bets(bets_batch)
                logging.info(f'action: apuesta_recibida | result: success | cantidad: {len(bets_batch)}')

        except OSError as e:
            logging.error(f'action: receive_message | result: fail | error: {e}')
        except ValueError as e:
            logging.error(f'action: receive_message | result: fail | error: {e}')
            protocol.send_error_code()
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
        