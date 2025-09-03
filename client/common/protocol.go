package common

import (
	"fmt"
	"strings"
)

const _SEPARATOR = "|"
const _MAX_BET_SIZE = 255
const _BET_RECEIVED = 1

type Protocol struct {
	socket *Socket
}

func NewProtocol(serverAddress string) (*Protocol, error) {
	socket, err := Connect(serverAddress)
	if err != nil {
		return nil, err
	}
	return &Protocol{socket: socket}, nil
}

func (proto *Protocol) Close() error {
	return proto.socket.Close()
}

func (proto *Protocol) serializeBet(bet *Bet) []byte {
	serialized := strings.Join([]string{
		bet.agency, bet.firstName, bet.lastName, bet.document, bet.birthday, bet.number,
	}, _SEPARATOR)
	return []byte(serialized)
}

func (proto *Protocol) SendBet(bet *Bet) error {
	serializedBet := proto.serializeBet(bet)

	if len(serializedBet) > _MAX_BET_SIZE {
		return fmt.Errorf("serialized bet is too long")
	}

	buf := make([]byte, 0, len(serializedBet)+1)
	buf = append(buf, byte(len(serializedBet)))
	buf = append(buf, serializedBet...)

	return proto.socket.SendAll(buf)
}

func (proto *Protocol) WaitConfirmation() error {
	buf, err := proto.socket.ReceiveAll(1)
	if err != nil {
		return err
	}
	if int(buf[0]) != _BET_RECEIVED {
		return fmt.Errorf("confirmation not received")
	}
	return nil
}