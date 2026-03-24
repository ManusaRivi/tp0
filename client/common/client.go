package common

import (
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/7574-sistemas-distribuidos/docker-compose-init/client/network"
	"github.com/7574-sistemas-distribuidos/docker-compose-init/client/repository"
	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("log")

type ClientStatus int

const (
	ClientStatusSendingBets ClientStatus = iota
	ClientStatusFinishedSendingBets
	ClientStatusRequestingWinners
	ClientStatusFinished
)

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
	config      ClientConfig
	skt         *network.Socket
	repository  *repository.Repository
	status      ClientStatus
	doneLooping chan struct{}
}

// NewClient Initializes a new client receiving the configuration
// as a parameter
func NewClient(config ClientConfig, repository *repository.Repository, doneLooping chan struct{}) *Client {
	client := &Client{
		config:      config,
		repository:  repository,
		status:      ClientStatusSendingBets,
		doneLooping: doneLooping,
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

func (c *Client) sendBetBatch(agencyID uint8) error {
	bets, isLastBatch, err := c.repository.FetchBets(c.config.BatchAmount)

	if err != nil {
		log.Errorf("action: obtener_apuestas | result: fail | client_id: %v | error: %v",
			c.config.ID,
			err,
		)
		return err
	}

	if len(bets) == 0 {
		c.status = ClientStatusFinishedSendingBets
	}

	// Create the connection to the server in every loop iteration.
	err = c.createClientSocket()
	if err != nil {
		return err
	}

	log.Debugf("action: obtener_apuestas | result: success | client_id: %v | batch_amount: %v",
		c.config.ID,
		len(bets),
	)

	sentBatch, betsProcessed, err := network.SendBetBatch(
		c.skt,
		agencyID,
		bets,
	)

	if err != nil {
		log.Errorf("action: apuesta_enviada | result: fail | error: %v",
			err,
		)
		c.skt.Close()
		return err
	}

	// If batch was not sent (maybe all bets were invalid, failed to parse) we close the connection,
	// advance the batch and continue with the next loop iteration, without waiting for a server response.
	if !sentBatch {
		c.skt.Close()
		c.repository.AdvanceBatch(betsProcessed)
		return nil
	}

	status, err := network.ReceiveAck(c.skt)
	if err != nil {
		log.Errorf("action: receive_message | result: fail | client_id: %v | error: %v",
			c.config.ID,
			err,
		)
		c.skt.Close()
		return err
	}

	if status != network.ServerAckStatusSuccess {
		log.Infof("action: apuesta_enviada | result: fail | status: %v",
			status,
		)
		c.skt.Close()
		return nil
	}

	log.Infof("action: apuesta_enviada | result: success")
	c.repository.AdvanceBatch(betsProcessed)
	if isLastBatch {
		log.Infof("action: loop_finished | result: success | client_id: %v", c.config.ID)
		c.status = ClientStatusFinishedSendingBets
	}
	c.skt.Close()
	return nil
}

func (c *Client) sendFinishedMessage(agencyID uint8) error {
	if err := c.createClientSocket(); err != nil {
		return err
	}

	if err := network.SendFinishedMessage(c.skt, agencyID); err != nil {
		c.skt.Close()
		log.Errorf("action: envio_finalizado | result: fail | client_id: %v | error: %v", agencyID, err)
		return err
	}

	status, err := network.ReceiveAck(c.skt)
	c.skt.Close()
	if err != nil {
		return err
	}

	if status != network.ServerAckStatusSuccess {
		return fmt.Errorf("unexpected FINISHED ACK status: %d", status)
	}

	c.status = ClientStatusRequestingWinners
	log.Infof("action: envio_finalizado | result: success")

	return nil
}

func (c *Client) requestWinners(agencyID uint8) error {
	if err := c.createClientSocket(); err != nil {
		return err
	}

	if err := network.SendWinnersRequestMessage(c.skt, agencyID); err != nil {
		c.skt.Close()
		log.Errorf("action: consulta_ganadores | result: fail | client_id: %v | error: %v", agencyID, err)
		return err
	}

	msgType, payload, err := network.ReceiveMessage(c.skt)
	c.skt.Close()
	if err != nil {
		return err
	}

	if msgType != network.MessageTypeWinnersResp {
		return fmt.Errorf("unexpected message type for winners response: %d", msgType)
	}

	dnis, err := network.ParseWinnersResponsePayload(payload)
	if err != nil {
		return err
	}

	if len(dnis) == 0 {
		log.Infof("action: consulta_ganadores | result: fail | status: still_processing_winners")
		return nil
	}

	c.status = ClientStatusFinished
	log.Infof("action: consulta_ganadores | result: success | cant_ganadores: %d", len(dnis))
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
		// Dispatch based on client status: sending bets, finished sending bets, requesting winners, finished.
		switch c.status {
		case ClientStatusSendingBets:
			c.sendBetBatch(uint8(agencyID))
		case ClientStatusFinishedSendingBets:
			c.sendFinishedMessage(uint8(agencyID))
		case ClientStatusRequestingWinners:
			c.requestWinners(uint8(agencyID))
		case ClientStatusFinished:
			c.doneLooping <- struct{}{}
			return
		}

		// Wait a time between sending one message and the next one
		time.Sleep(c.config.LoopPeriod)

	}
	log.Infof("action: loop_finished | result: success | client_id: %v", c.config.ID)
	c.doneLooping <- struct{}{}
}

// Close Closes the client connection.
func (c *Client) Close() {
	if c.skt != nil {
		c.skt.Close()
	}
}
