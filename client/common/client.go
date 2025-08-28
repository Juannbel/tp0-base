package common

import (
	"bufio"
	"fmt"
	"net"
	"time"

	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("log")

// ClientConfig Configuration used by the client
type ClientConfig struct {
	ID            string
	ServerAddress string
	LoopAmount    int
	LoopPeriod    time.Duration
}

// Client Entity that encapsulates how
type Client struct {
	config ClientConfig
	conn   net.Conn
	keepRunning bool
	stopChannel chan struct{}
}

// NewClient Initializes a new client receiving the configuration
// as a parameter
func NewClient(config ClientConfig) *Client {
	client := &Client{
		config:     config,
		keepRunning: true,
		stopChannel: make(chan struct{}),
	}
	return client
}

// CreateClientSocket Initializes client socket. In case of
// failure, error is printed in stdout/stderr and exit 1
// is returned
func (c *Client) createClientSocket() error {
	conn, err := net.Dial("tcp", c.config.ServerAddress)
	if err != nil {
		log.Criticalf(
			"action: connect | result: fail | client_id: %v | error: %v",
			c.config.ID,
			err,
		)
	}
	c.conn = conn
	return nil
}

// StartClientLoop Send messages to the client until some time threshold is met
func (c *Client) StartClientLoop() {
	defer c.cleanup()
	
	// There is an autoincremental msgID to identify every message sent
	// Messages if the message amount threshold has not been surpassed
	for msgID := 1; msgID <= c.config.LoopAmount && c.keepRunning; msgID++ {
		// Create the connection the server in every loop iteration. Send an
		c.createClientSocket()

		// TODO: Modify the send to avoid short-write
		fmt.Fprintf(
			c.conn,
			"[CLIENT %v] Message N°%v\n",
			c.config.ID,
			msgID,
		)
		msg, err := bufio.NewReader(c.conn).ReadString('\n')
		c.conn.Close()

		if err != nil {
			log.Errorf("action: receive_message | result: fail | client_id: %v | error: %v",
				c.config.ID,
				err,
			)
			return
		}

		log.Infof("action: receive_message | result: success | client_id: %v | msg: %v",
			c.config.ID,
			msg,
		)

		select {
			// Wait a time between sending one message and the next one
			case <- time.After(c.config.LoopPeriod):
			case <- c.stopChannel: // el keepRunning en false lo saca del loop
				continue
		}
	}

	if !c.keepRunning {
		return
	}

	log.Infof("action: loop_finished | result: success | client_id: %v", c.config.ID)
	c.Stop()
}

func (c *Client) Stop() {
	if c.keepRunning {
		c.keepRunning = false
		close(c.stopChannel)
		log.Infof("action: client_stopped | result: success | client_id: %v", c.config.ID)
	} else {
		log.Fatalf("action: client_stop | result: fail | client_id: %v | error: client_already_stopped", c.config.ID)
	}
}

func (c *Client) cleanup() {
	if c.conn != nil {
		c.conn.Close()
		log.Infof("action: client_connection_closed | result: success | client_id: %v", c.config.ID)
	}

	log.Infof("action: client_cleanup | result: success | client_id: %v", c.config.ID)
}