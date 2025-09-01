import socket
import logging
from common.protocol import Protocol, SENDING_BETS, REQUEST_RESULTS
from common.utils import store_bets, load_bets, has_won

class Server:
    def __init__(self, port, listen_backlog, number_of_agencies):
        # Initialize server socket
        self._server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self._server_socket.bind(('', port))
        self._server_socket.listen(listen_backlog)
        self._keep_running = True
        self._number_of_agencies = number_of_agencies
        self._processed_agencies = 0
        self._winners = {}
        
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
            
            action = protocol.receive_action()
            if action == SENDING_BETS:
                self.__handle_sending_bets(protocol)
                if self._processed_agencies == self._number_of_agencies:
                    logging.debug('action: all_agencies_processed | result: success')
                    self.__perform_raffle()

            elif action == REQUEST_RESULTS:
                self.__handle_request_results(protocol)

        except OSError as e:
            logging.error(f'action: receive_message | result: fail | error: {e}')
        except ValueError as e:
            logging.error(f'action: receive_message | result: fail | error: {e}')
            protocol.send_error_code()
        finally:
            protocol.close()

    def __handle_sending_bets(self, protocol):
        logging.debug('action: receive_bets | result: in_progress')

        while self._keep_running:
            bets_batch = protocol.receive_bets_batch()

            if not bets_batch:
                logging.debug('action: receive_bets | result: success | info: no more bets')
                self._processed_agencies += 1
                return

            protocol.confirm_reception()

            store_bets(bets_batch)
            logging.info(f'action: apuesta_recibida | result: success | cantidad: {len(bets_batch)}')

    def __handle_request_results(self, protocol):
        agency = protocol.receive_agency_id()
        if self._processed_agencies < self._number_of_agencies:
            protocol.send_results_not_ready()
            logging.debug(f'action: request_results | result: fail | agency: {agency} | info: results not ready')
            return

        winners = self._winners.get(agency, [])
        protocol.send_winners(winners)
        logging.debug(f'action: request_results | result: success | agency: {agency}')

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
    
    def __perform_raffle(self):
        for bet in load_bets():
            if has_won(bet):
                if bet.agency not in self._winners:
                    self._winners[bet.agency] = []

                self._winners[bet.agency].append(bet.document)
        
        logging.info('action: sorteo | result: success')
                
    def stop(self, signum, frame):
        """
        Stop the server gracefully
        """
        logging.info('action: stop_server | result: in_progress')
        self._keep_running = False
        self._server_socket.close()
        logging.info('action: close_server_socket | result: success')
        