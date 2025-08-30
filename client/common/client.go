package common

import (
	"bufio"
	"os"

	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("log")

// ClientConfig Configuration used by the client
type ClientConfig struct {
	ID            string
	ServerAddress string
	batchAmount   int
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

func (c *Client) Start() error {
	defer c.cleanup()
	csvFile, err := os.Open("/agency.csv")
	if err != nil {
		log.Criticalf(
			"action: open_csv | result: fail | client_id: %v | error: %v",
			c.config.ID,
			err,
		)
		return err
	}
	defer csvFile.Close()

	csvReader := bufio.NewScanner(csvFile)
	batchGenerator := NewBatchGenerator(c.config.ID, c.config.batchAmount, csvReader, c.proto.GetBetSize)

	for {
		batch, err := batchGenerator.GetNextBatch()

		if err != nil {
			log.Errorf("action: read_batch | result: fail | client_id: %v | error: %v",
				c.config.ID,
				err,
			)
			return err
		}

		if len(batch) == 0 {
			break
		}

		if err := c.proto.SendBatch(batch); err != nil {
			log.Errorf("action: apuesta_enviada | result: fail | client_id: %v | error: %v",
				c.config.ID,
				err,
			)
			return err
		}

		if err := c.proto.WaitConfirmation(); err != nil {
			log.Errorf("action: wait_confirmation | result: fail | client_id: %v | error: %v",
				c.config.ID,
				err,
			)
			return err
		}
		
		log.Infof("action: apuesta_enviada | result: success | client_id: %v | cantidad: %v",
			c.config.ID,
			len(batch),
		)
	}

	c.proto.InformCompletion()
	return nil
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