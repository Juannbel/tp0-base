import socket
import logging
from common.protocol import Protocol, SENDING_BETS, REQUEST_RESULTS
from common.utils import store_bets, load_bets, has_won
from threading import Thread, Lock

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
        self._client_handlers = set()
        self._lock = Lock()

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
                protocol = Protocol(client_sock)
                thread = Thread(target=self.__handle_client_connection, args=(protocol,))
                self._client_handlers.add((thread, protocol))
        
                thread.start()
            
            self.__reap_dead()

        logging.info('action: stop_server | result: success')

    def __handle_client_connection(self, protocol):
        try:
            action = protocol.receive_action()
            if action == SENDING_BETS:
                self.__handle_sending_bets(protocol)

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
                with self._lock:
                    logging.debug('action: receive_bets | result: success | info: no more bets')
                    self._processed_agencies += 1
                    
                    if self._processed_agencies == self._number_of_agencies:
                        logging.debug('action: all_agencies_processed | result: success')
                        self.__perform_raffle()
                
                return

            protocol.confirm_reception()
            
            with self._lock:
                store_bets(bets_batch)
                
            logging.info(f'action: apuesta_recibida | result: success | cantidad: {len(bets_batch)}')

    def __handle_request_results(self, protocol):
        agency = protocol.receive_agency_id()
        winners = []
        ready = False
        with self._lock:
            if self._processed_agencies == self._number_of_agencies:
                winners = self._winners.get(agency, [])
                ready = True

        if ready:
            protocol.send_winners(winners)  
            logging.debug(f'action: request_results | result: success | agency: {agency}')
        else:
            protocol.send_results_not_ready()

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
        """
        Perform the raffle, only called with the lock held
        """
        for bet in load_bets():
            if has_won(bet):
                if bet.agency not in self._winners:
                    self._winners[bet.agency] = []

                self._winners[bet.agency].append(bet.document)
        
        logging.info('action: sorteo | result: success')
        
    def __reap_dead(self):
        to_remove = []
        for thread, protocol in self._client_handlers:
            if not thread.is_alive():
                to_remove.append((thread, protocol))
                
        for thread, protocol in to_remove:
            self._client_handlers.remove((thread, protocol))
            protocol.close()

    def stop(self, signum, frame):
        """
        Stop the server gracefully
        """
        logging.info('action: stop_server | result: in_progress')
        for thread, protocol in self._client_handlers:
            try:
                protocol.close()
                thread.join()
            except Exception as e:
                logging.error(f'action: stop_server | result: thread_exception | error: {e}')

        self._keep_running = False
        self._server_socket.close()
        logging.info('action: close_server_socket | result: success')
        