package handlers

import (
	"errors"

	"github.com/CodeCollaborate/Server/modules/config"
	"github.com/CodeCollaborate/Server/modules/datahandling"
	"github.com/CodeCollaborate/Server/modules/dbfs"
	"github.com/CodeCollaborate/Server/modules/rabbitmq"
	"github.com/CodeCollaborate/Server/utils"
)

// server only has one worker
var workerCfg *rabbitmq.AMQPPubSubCfg

const workerName string = "datahandling_worker"
const workerOutboundQueueBufferSize int = 512

// StartWorker initializes the worker which talks with RabbitMQ for this server
func StartWorker(dbfsImpl dbfs.DBFS, prefetchCount int) *rabbitmq.AMQPPubSubCfg {
	if workerCfg != nil {
		// Worker already initialized, restarting
		workerCfg.Control.Shutdown()
		workerCfg = nil
	}

	cfg := config.GetConfig()

	pubCfg := rabbitmq.NewPubConfig(func(msg rabbitmq.AMQPMessage) {
		// do nothing (for now?)
		msg.ErrHandler()
	}, workerOutboundQueueBufferSize)

	subCfg := &rabbitmq.AMQPSubCfg{
		QueueName:     workerName,
		Keys:          []string{},
		IsWorkQueue:   true,
		PrefetchCount: prefetchCount,
	}

	workerCfg = rabbitmq.NewAMQPPubSubCfg(cfg.ServerConfig.Name, pubCfg, subCfg)

	subCfg.HandleMessageFunc = workerMessageHandler(dbfsImpl, workerCfg)

	go func() {
		// auto-restart on failure
		for !workerCfg.Control.HasExited() {
			err := rabbitmq.RunPublisher(workerCfg)
			if err != nil {
				utils.LogError("Worker publisher error encountered. Exiting", err, nil)
			}
		}
	}()
	go func() {
		// auto-restart on failure
		for !workerCfg.Control.HasExited() {
			err := rabbitmq.RunSubscriber(workerCfg)
			if err != nil {
				utils.LogError("Worker subscriber error encountered. Exiting", err, nil)
			}
		}
	}()

	workerCfg.Control.Ready.Wait()
	return workerCfg
}

// WorkerEnqueue takes messages that need to be processed and sends them to RabbitMQ to be assigned to a worker
func WorkerEnqueue(message []byte, wsID uint64) error {
	// can't naturally get wsID's of 0
	if wsID == 0 {
		return errors.New("invalid websocketID given to worker")
	}

	msg := rabbitmq.AMQPMessage{
		Headers: map[string]interface{}{
			"Origin": rabbitmq.LocalWebsocketName(wsID),
		},
		RoutingKey:  workerName,
		ContentType: rabbitmq.ContentTypeWork,
		Persistent:  false,
		Message:     message,
	}

	select {
	case workerCfg.PubCfg.Messages <- msg:
	default:
		err := errors.New("Channel buffer full")
		utils.LogError("Worker message queue full, failed to add new message", err, utils.LogFields{
			"AMQP Message": msg,
		})
		return err
	}

	return nil
}

func workerMessageHandler(dbfsImpl dbfs.DBFS, cfg *rabbitmq.AMQPPubSubCfg) func(rabbitmq.AMQPMessage) error {
	// have 1 dh per worker
	dh := datahandling.DataHandler{
		MessageChan: cfg.PubCfg.Messages,
		Db:          dbfsImpl,
	}

	return func(msg rabbitmq.AMQPMessage) error {
		switch msg.ContentType {
		case rabbitmq.ContentTypeWork:
			// If notification with self as origin, early-out; ignore our own notifications.
			wsOrigin, ok := msg.Headers["Origin"]
			if !ok {
				err := errors.New("Unnown message origin")
				utils.LogError("Worker encountered ", err, utils.LogFields{
					"Message Headers": msg.Headers, // NOTE: message body could contain passwords
				})
				return err
			}

			go dh.Handle(msg.Message, wsOrigin.(string), msg.Ack)
			return nil
		default:
			err := errors.New("Unnable to process RabbitMQ message type")
			utils.LogError("not-work given to worker", err, utils.LogFields{
				"Message Headers": msg.Headers,
				"Message Body":    string(msg.Message),
			})
			return err
		}
	}
}