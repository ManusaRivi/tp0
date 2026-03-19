package common

import (
	"net"
	"time"

	"github.com/7574-sistemas-distribuidos/docker-compose-init/client/network"
	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("log")

type AgencyData struct {
	FirstName  string
	LastName   string
	DNI        uint32
	BirthYear  uint16
	BirthMonth uint8
	BirthDay   uint8
	Number     uint32
}

// ClientConfig Configuration used by the client
type ClientConfig struct {
	ID            string
	ServerAddress string
	LoopAmount    int
	LoopPeriod    time.Duration
	AgencyData    AgencyData
}

// Client Entity that encapsulates how
type Client struct {
	config ClientConfig
	skt    *network.Socket
}

// NewClient Initializes a new client receiving the configuration
// as a parameter
func NewClient(config ClientConfig) *Client {
	client := &Client{
		config: config,
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
		return err
	}
	c.skt = network.NewSocket(conn)
	return nil
}

// StartClientLoop Send messages to the client until some time threshold is met
func (c *Client) StartClientLoop() {
	// There is an autoincremental msgID to identify every message sent
	// Messages if the message amount threshold has not been surpassed
	for msgID := 1; msgID <= c.config.LoopAmount; msgID++ {
		// Create the connection the server in every loop iteration. Send an
		err := c.createClientSocket()

		if err != nil {
			return
		}

		err = network.SendBetMessage(
			c.skt,
			c.config.AgencyData.FirstName,
			c.config.AgencyData.LastName,
			c.config.AgencyData.DNI,
			c.config.AgencyData.BirthYear,
			c.config.AgencyData.BirthMonth,
			c.config.AgencyData.BirthDay,
			c.config.AgencyData.Number,
		)

		if err != nil {
			log.Errorf("action: apuesta_enviada | result: fail | dni: %v | numero: %v | error: $v",
				c.config.AgencyData.DNI,
				c.config.AgencyData.Number,
				err,
			)
			c.skt.Close()
			return
		}

		status, err := network.ReceiveServerMessage(c.skt)

		switch status {
		case network.ServerMessageStatusFailure:
			log.Infof("action: apuesta_enviada | result: fail | dni: %v | numero: %v",
				c.config.AgencyData.DNI,
				c.config.AgencyData.Number,
			)
		case network.ServerMessageStatusSuccess:
			log.Infof("action: apuesta_enviada | result: success | dni: %v | numero: %v",
				c.config.AgencyData.DNI,
				c.config.AgencyData.Number,
			)
		}
		c.skt.Close()

		if err != nil {
			log.Errorf("action: receive_message | result: fail | client_id: %v | error: %v",
				c.config.ID,
				err,
			)
			return
		}

		// Wait a time between sending one message and the next one
		time.Sleep(c.config.LoopPeriod)

	}
	log.Infof("action: loop_finished | result: success | client_id: %v", c.config.ID)
}

// Close Closes the client connection.
func (c *Client) Close() {
	if c.skt != nil {
		c.skt.Close()
	}
}
