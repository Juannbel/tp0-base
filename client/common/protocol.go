package common

import (
	"encoding/binary"
	"fmt"
	"strings"
)

const _BET_SEPARATOR = "|"
const _BATCH_SEPARATOR = "#"
// entre cada campo va un "|" y un "#" al final
const _SEPARATORS_PER_BET = 6
const _BATCH_RECEIVED = 1
const _ERROR_CODE = 2

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

func (proto *Protocol) SendBatch(batch []*Bet) error {
	serializedBets := make([]string, 0, len(batch))
	for _, bet := range batch {
		serializedBet := proto.serialize(bet)
		serializedBets = append(serializedBets, serializedBet)
	}

	serializedBatch := strings.Join(serializedBets, _BATCH_SEPARATOR)
	totalSize := uint16(len(serializedBatch))

	buf := proto.uint16ToBytes(totalSize)
	buf = append(buf, serializedBatch...)

	return proto.socket.SendAll(buf)
}

func (proto *Protocol) WaitConfirmation() error {
	action, err := proto.receiveAction()
	if err != nil {
		return err
	}

	switch action {
		case _BATCH_RECEIVED:
			return nil
		case _ERROR_CODE:
			return fmt.Errorf("error received from server")
		default:
			return fmt.Errorf("unexpected code received from server: %d", action)
	}
}

func (proto *Protocol) InformCompletion() error {
	buf := []byte{0, 0}
	return proto.socket.SendAll(buf)
}

func (proto *Protocol) uint16ToBytes(value uint16) []byte {
	buf := make([]byte, 2)
	binary.BigEndian.PutUint16(buf, value)
	return buf
}

func (proto *Protocol) receiveAction() (int, error) {
	buf, err := proto.socket.ReceiveAll(1)
	if err != nil {
		return 0, err
	}
	return int(buf[0]), nil
}