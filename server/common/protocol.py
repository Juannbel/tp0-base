import logging
from common.socket import Socket
from common.utils import Bet

BET_SEPARATOR = '|'
BATCH_SEPARATOR = '#'
WINNER_SEPARATOR = '$'

BET_PARTS = 6

SENDING_BETS = b'\x00'
BATCH_RECEIVED = b'\x01'
REQUEST_RESULTS = b'\x02'
RESULTS_NOT_READY = b'\x03'
SENDING_RESULTS = b'\x04'
ERROR_CODE = b'\x05'

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
        batch_length = self.__receive_uint16()
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
    
    def receive_action(self):
        return self._sock.recvall(1)
    
    def receive_agency_id(self):
        return ord(self._sock.recvall(1))

    def send_winners(self, winners):
        buf = b''
        serialized_winners = WINNER_SEPARATOR.join(winners).encode('utf-8')
        winners_len = len(serialized_winners)

        buf += SENDING_RESULTS
        buf += winners_len.to_bytes(2, byteorder='big')
        buf += serialized_winners

        self._sock.sendall(buf)
    
    def __receive_uint16(self):
        data = self._sock.recvall(2)
        return int.from_bytes(data, byteorder='big', signed=False)

    def send_results_not_ready(self):
        self._sock.sendall(RESULTS_NOT_READY)

    def confirm_reception(self):
        self._sock.sendall(BATCH_RECEIVED)

    def send_error_code(self):
        self._sock.sendall(ERROR_CODE)

    def close(self):
        self._sock.close()