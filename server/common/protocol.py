import logging
from common.socket import Socket
from common.utils import Bet

BET_SEPARATOR = '|'
BATCH_SEPARATOR = '#'
BET_PARTS = 6
BATCH_RECEIVED = b'\x01'
ERROR_CODE = b'\x02'

class Protocol:
    def __init__(self, sock):
        self._sock = Socket(sock)

    def deserialize_bet(self, serialized):
        parts = serialized.split(BET_SEPARATOR)
        if len(parts) != BET_PARTS:
            return None

        agency, first_name, last_name, document, birthday, number = parts
        return Bet(agency, first_name, last_name, document, birthday, number)

    def receive_bets_batch(self):
        batch_length_be = self._sock.recvall(2)
        batch_length = int.from_bytes(batch_length_be, byteorder='big', signed=False)
        
        
        if batch_length == 0:
            return []
        
        batch_data = self._sock.recvall(batch_length).decode('utf-8')

        bets = []
        for serialized_bet in batch_data.split(BATCH_SEPARATOR):
            bet = self.deserialize_bet(serialized_bet)
            if bet is None:
                logging.error(f'action: apuesta_recibida | result: fail | cantidad: {batch_length}')
                raise ValueError('Invalid bet format')
            
            bets.append(bet)
            
        return bets

    def confirm_reception(self):
        self._sock.sendall(BATCH_RECEIVED)

    def send_error_code(self):
        self._sock.sendall(ERROR_CODE)

    def close(self):
        self._sock.close()