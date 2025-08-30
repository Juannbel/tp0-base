package common

import (
	"bufio"
	"fmt"
)

const _MAX_BATCH_SIZE = 1024 * 8

type BatchGenerator struct {
	agency    string
	pendingBet *Bet
	csvReader  *bufio.Scanner
	batchAmount int
	betSize     func(b *Bet) int
}

func NewBatchGenerator(agency string, batchAmount int, csvReader *bufio.Scanner, betSize func(b *Bet) int) *BatchGenerator {
	return &BatchGenerator{
		agency:    agency,
		pendingBet: nil,
		csvReader:  csvReader,
		batchAmount: batchAmount,
		betSize:    betSize,
	}
}

func (bg *BatchGenerator) GetNextBatch() ([]*Bet, error) {
	batch := make([]*Bet, 0)
	serializedSize := 0

	if bg.pendingBet != nil {
		batch = append(batch, bg.pendingBet)
		serializedSize = bg.betSize(bg.pendingBet)
		bg.pendingBet = nil
	}

	for len(batch) < bg.batchAmount {
		if !bg.csvReader.Scan() {
			break
		}

		line := bg.csvReader.Text()
		bet := CreateBetFromCSVLine(bg.agency,line)
		if bet == nil {
			return nil, fmt.Errorf("error parsing csv line")
		}

		if bg.betSize(bet) + serializedSize > _MAX_BATCH_SIZE {
			bg.pendingBet = bet
			break
		} else {
			batch = append(batch, bet)
			serializedSize += bg.betSize(bet)
		}
	}

	return batch, nil
}