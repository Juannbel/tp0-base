package common

import (
	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("log")

// ClientConfig Configuration used by the client
type ClientConfig struct {
	ID            string
	ServerAddress string
}

// Client Entity that encapsulates how
type Client struct {
	config ClientConfig
	proto  *Protocol
	stopChannel chan struct{}
}

// NewClient Initializes a new client receiving the configuration
// as a parameter
func NewClient(config ClientConfig) *Client {
	proto, err := NewProtocol(config.ServerAddress)
	if err != nil {
		log.Criticalf(
			"action: connect | result: fail | client_id: %v | error: %v",
			config.ID,
			err,
		)
		return nil
	}

	return &Client{
		config:     config,
		proto:     proto,
		stopChannel: make(chan struct{}),
	}
}

// Creates the bet, sends it, and wait for the confirmation
func (c *Client) Start() {
	defer c.cleanup()

	bet := CreateBetFromEnv()
	if err := c.proto.SendBet(bet); err != nil {
		log.Errorf("action: send_bet | result: fail | client_id: %v | error: %v",
			c.config.ID,
			err,
		)
		return
	}

	if err := c.proto.WaitConfirmation(); err != nil {
		log.Errorf("action: wait_confirmation | result: fail | client_id: %v | error: %v",
			c.config.ID,
			err,
		)
		return
	}

	log.Infof("action: apuesta_enviada | result: success | dni: %v | numero: %v",
		bet.document,
		bet.number,
	)
}

func (c *Client) Stop() {
	c.proto.Close()
	log.Infof("action: client_stopped | result: success | client_id: %v", c.config.ID)
}

func (c *Client) cleanup() {
	c.proto.Close()
	log.Infof("action: client_connection_closed | result: success | client_id: %v", c.config.ID)

	log.Infof("action: client_cleanup | result: success | client_id: %v", c.config.ID)
}