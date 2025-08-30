package common

import (
	"strings"
)

type Bet struct {
	agency  string
	firstName string
	lastName  string
	document string
	birthday string
	number string
}

func CreateBetFromCSVLine(agency string, line string) *Bet {
	parts := strings.Split(line, ",")
	if len(parts) < 5 {
		return nil
	}
	return &Bet{
		agency:  agency,
		firstName: parts[0],
		lastName:  parts[1],
		document: parts[2],
		birthday: parts[3],
		number:   parts[4],
	}
}