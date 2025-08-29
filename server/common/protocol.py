from common.socket import Socket
from common.utils import Bet

SEPARATOR = '|'
BET_RECEIVED = b'\x01'

class Protocol:
    def __init__(self, sock):
        self._sock = Socket(sock)
        
    def receive_bet(self):
        bet_length = self._sock.recvall(1)
        bet_data = self._sock.recvall(ord(bet_length))

        agency, first_name, last_name, document, birthday, number = bet_data.decode('utf-8').split(SEPARATOR)
        return Bet(agency, first_name, last_name, document, birthday, number)

    def confirm_reception(self):
        self._sock.sendall(BET_RECEIVED)

    def close(self):
        self._sock.close()