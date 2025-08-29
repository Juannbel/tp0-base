package common

import (
	"os"
)

type Bet struct {
	agency  string
	firstName string
	lastName  string
	document string
	birthday string
	number string
}

func CreateBetFromEnv() *Bet {
	return &Bet{
		agency:  os.Getenv("CLI_ID"),
		firstName: os.Getenv("NOMBRE"),
		lastName:  os.Getenv("APELLIDO"),
		document: os.Getenv("DOCUMENTO"),
		birthday: os.Getenv("NACIMIENTO"),
		number:   os.Getenv("NUMERO"),
	}
}