package common

import (
	"net"
	"strconv"
	"time"

	"github.com/7574-sistemas-distribuidos/docker-compose-init/client/network"
	"github.com/7574-sistemas-distribuidos/docker-compose-init/client/repository"
	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("log")

// ClientConfig Configuration used by the client
type ClientConfig struct {
	ID            string
	ServerAddress string
	LoopAmount    int
	LoopPeriod    time.Duration
	BatchAmount   int
}

// Client Entity that encapsulates how
type Client struct {
	config     ClientConfig
	skt        *network.Socket
	repository *repository.Repository
}

// NewClient Initializes a new client receiving the configuration
// as a parameter
func NewClient(config ClientConfig, repository *repository.Repository) *Client {
	client := &Client{
		config:     config,
		repository: repository,
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
	agencyID, err := strconv.ParseUint(c.config.ID, 10, 8)
	if err != nil {
		log.Errorf("action: parse_client_id | result: fail | client_id: %v | error: %v",
			c.config.ID,
			err,
		)
		return
	}

	// There is an autoincremental msgID to identify every message sent
	// Messages if the message amount threshold has not been surpassed
	for msgID := 1; msgID <= c.config.LoopAmount; msgID++ {
		// Create the connection the server in every loop iteration. Send an
		err := c.createClientSocket()

		if err != nil {
			return
		}

		bets, err := c.repository.FetchBets(c.config.BatchAmount)

		if err != nil {
			log.Errorf("action: obtener_apuestas | result: fail | client_id: %v | batch_amount: %v | error: %v",
				c.config.ID,
				len(bets),
				err,
			)
			c.skt.Close()
			return
		}

		log.Debugf("action: obtener_apuestas | result: success | client_id: %v | batch_amount: %v",
			c.config.ID,
			len(bets),
		)

		err = network.SendBetBatch(
			c.skt,
			uint8(agencyID),
			bets,
		)

		if err != nil {
			log.Errorf("action: apuesta_enviada | result: fail | error: %v",
				err,
			)
			c.skt.Close()
			return
		}

		status, err := network.ReceiveServerMessage(c.skt)

		switch status {
		case network.ServerMessageStatusFailure:
			log.Infof("action: apuesta_enviada | result: fail")
		case network.ServerMessageStatusSuccess:
			log.Infof("action: apuesta_enviada | result: success")
		default:
			log.Warningf("action: apuesta_enviada | result: fail | status: %v",
				status,
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
