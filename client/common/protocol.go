package common

import (
	"encoding/binary"
	"fmt"
	"strings"
)

const _BET_SEPARATOR = "|"
const _BATCH_SEPARATOR = "#"
const _WINNER_SEPARATOR = "$"
// entre cada campo va un "|" y un "#" al final
const _SEPARATORS_PER_BET = 6

const _SENDING_BETS = 0
const _BATCH_RECEIVED = 1
const _REQUEST_RESULTS = 2
const _RESULTS_NOT_READY = 3
const _SENDING_RESULTS = 4
const _ERROR_CODE = 5

type Protocol struct {
	socket *Socket
	GetBetSize func(b *Bet) int
}

func NewProtocol(serverAddress string) (*Protocol, error) {
	socket, err := Connect(serverAddress)
	if err != nil {
		return nil, err
	}

	betSize := func(b *Bet) int {
		return len(b.agency) +
			len(b.firstName) + 
			len(b.lastName) + 
			len(b.document) + 
			len(b.birthday) + 
			len(b.number) + _SEPARATORS_PER_BET
	}
	return &Protocol{socket: socket, GetBetSize: betSize}, nil
}

func (proto *Protocol) Close() error {
	return proto.socket.Close()
}

func (proto *Protocol) serialize(bet *Bet) string {
	serialized := strings.Join([]string{
		bet.agency, bet.firstName, bet.lastName, bet.document, bet.birthday, bet.number,
	}, _BET_SEPARATOR)
	return serialized
}

func (proto *Protocol) StartSendingBets() error {
	buf := []byte{_SENDING_BETS}
	return proto.socket.SendAll(buf)
}

func (proto *Protocol) RequestResults(agencyId int) ([]string, error) {
	buf := []byte{_REQUEST_RESULTS, byte(agencyId)}
	err := proto.socket.SendAll(buf)
	if err != nil {
		return nil, err
	}

	action, err := proto.receiveAction()
	if err != nil {
		return nil, err
	}

	if action == _RESULTS_NOT_READY {
		return nil, nil
	} else if action == _SENDING_RESULTS {
		return proto.receiveWinners()
	} else {
		return nil, fmt.Errorf("unexpected code received from server: %d", action)
	}
}

func (proto *Protocol) receiveWinners() ([]string, error) {
	winnersLen, err := proto.receiveUint16()
	if err != nil {
		return nil, err
	}

	serializedWinners, err := proto.socket.ReceiveAll(int(winnersLen))
	if err != nil {
		return nil, err
	}

	return strings.Split(string(serializedWinners), _WINNER_SEPARATOR), nil
}

func (proto *Protocol) SendBatch(batch []*Bet) error {
	serializedBets := make([]string, 0, len(batch))
	for _, bet := range batch {
		serializedBet := proto.serialize(bet)
		serializedBets = append(serializedBets, serializedBet)
	}

	serializedBatch := strings.Join(serializedBets, _BATCH_SEPARATOR)
	totalSize := len(serializedBatch)

	buf := make([]byte, 2+totalSize)
	binary.BigEndian.PutUint16(buf[:2], uint16(totalSize))
	copy(buf[2:], serializedBatch)

	return proto.socket.SendAll(buf)
}

func (proto *Protocol) WaitConfirmation() error {
	action, err := proto.receiveAction()
	if err != nil {
		return err
	}

	if action == _BATCH_RECEIVED {
		return nil
	} else if action == _ERROR_CODE {
		return fmt.Errorf("error received from server")
	} else {
		return fmt.Errorf("unexpected code received from server: %d", action)
	}
}

func (proto *Protocol) InformCompletion() error {
	buf := []byte{0, 0}
	return proto.socket.SendAll(buf)
}

func (proto *Protocol) receiveAction() (int, error) {
	buf, err := proto.socket.ReceiveAll(1)
	if err != nil {
		return 0, err
	}
	return int(buf[0]), nil
}

func (proto *Protocol) receiveUint16() (uint16, error) {
	buf, err := proto.socket.ReceiveAll(2)
	if err != nil {
		return 0, err
	}

	return binary.BigEndian.Uint16(buf), nil
}