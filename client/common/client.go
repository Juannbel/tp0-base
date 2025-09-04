package common

import (
	"bufio"
	"os"
	"strconv"
	"time"

	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("log")
const _SECONDS_TO_REASK = 1

// ClientConfig Configuration used by the client
type ClientConfig struct {
	ID            string
	ServerAddress string
	BatchAmount   int
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
	client := &Client{
		config:     config,
		stopChannel: make(chan struct{}),
	}

	if err := client.connectToServer(); err != nil {
		return nil
	}

	return client
}

func (c *Client) Start() {
	defer c.cleanup()

	if err := c.sendAllBets(); err != nil {
		return
	}

	c.proto.Close()

	c.waitWinners()
}

func (c *Client) sendAllBets() error {
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
	batchGenerator := NewBatchGenerator(c.config.ID, c.config.BatchAmount, csvReader, c.proto.GetBetSize)
	
	if err := c.proto.StartSendingBets(); err != nil {
		log.Criticalf(
			"action: start_sending_bets | result: fail | client_id: %v | error: %v",
			c.config.ID,
			err,
		)
		return err
	}

	for {
		sentBets, err := c.generateAndSendBatch(batchGenerator)
		if err != nil {
			return err
		}

		if sentBets == 0 {
			// All bets sent
			break
		}

		if err := c.proto.WaitConfirmation(); err != nil {
			log.Errorf("action: wait_confirmation | result: fail | client_id: %v | error: %v",
				c.config.ID,
				err,
			)
			return err
		}

		log.Debugf("action: apuesta_enviada | result: success | cantidad: %v",
			sentBets,
		)
	}

	c.proto.InformCompletion()
	return nil
}

func (c *Client) generateAndSendBatch(batchGenerator *BatchGenerator) (int, error) {
	batch, err := batchGenerator.GetNextBatch()
	if err != nil {
		log.Errorf("action: read_batch | result: fail | client_id: %v | error: %v",
			c.config.ID,
			err,
		)
		return 0, err
	}

	if len(batch) == 0 {
		return 0, nil
	}

	if err := c.proto.SendBatch(batch); err != nil {
		log.Errorf("action: apuesta_enviada | result: fail | client_id: %v | error: %v",
			c.config.ID,
			err,
		)
		return 0, err
	}

	return len(batch), nil
}

func (c *Client) connectToServer() error {
	proto, err := NewProtocol(c.config.ServerAddress)
	if err != nil {
		log.Criticalf(
			"action: connect | result: fail | client_id: %v | error: %v",
			c.config.ID,
			err,
		)
		return err
	}
	c.proto = proto
	return nil
}

func (c *Client) waitWinners() error {
	agencyId, _ := strconv.Atoi(c.config.ID)

	for {
		if err := c.connectToServer(); err != nil {
			log.Criticalf(
				"action: connect | result: fail | client_id: %v | error: %v",
				c.config.ID,
				err,
			)
			return err
		}

		winners, err := c.proto.RequestResults(agencyId)
		if err != nil {
			log.Criticalf(
				"action: consulta_ganadores | result: fail | client_id: %v | error: %v",
				c.config.ID,
				err,
			)
			return err
		}

		if winners != nil {
			log.Infof("action: consulta_ganadores | result: success | cant_ganadores: %v", len(winners))
			for i, winner := range winners {
				log.Debugf("action: consulta_ganadores | result: winner_%d | document: %v", i, winner)
			}
			return nil
		}

		c.proto.Close()

		select {
		case <-c.stopChannel:
			return nil
		case <- time.After(_SECONDS_TO_REASK * time.Second):
		}
	}
}
func (c *Client) Stop() {
	c.proto.Close()
	close(c.stopChannel)
	log.Infof("action: client_stopped | result: success | client_id: %v", c.config.ID)
}

func (c *Client) cleanup() {
	c.proto.Close()
	log.Infof("action: client_connection_closed | result: success | client_id: %v", c.config.ID)

	log.Infof("action: client_cleanup | result: success | client_id: %v", c.config.ID)
}