package rabbitmq

import (
	"crypto/tls"
	"fmt"
	"os"

	"github.com/CodeCollaborate/Server/modules/config"
	"github.com/CodeCollaborate/Server/utils"
)

/**
 * Configuration structures and variables for RabbitMQ.
 */

// Gets the hostname of this machine, for use in QueueName()
var hostname, _ = os.Hostname()

// AMQPConnCfg represents the settings needed to create a new connection, and initialize the required exchanges.
type AMQPConnCfg struct {
	config.ConnCfg
	Exchanges []AMQPExchCfg
	TLSConfig *tls.Config
	Control   *utils.Control
}

// ConnectionString returns the connection string, using amqps:// if TLSConfig has been set, amqp:// otherwise.
func (cfg AMQPConnCfg) ConnectionString() string {
	if cfg.TLSConfig != nil {
		return fmt.Sprintf("amqps://%s:%s@%s:%d/", cfg.Username, cfg.Password, cfg.Host, cfg.Port)
	}
	return fmt.Sprintf("amqp://%s:%s@%s:%d/", cfg.Username, cfg.Password, cfg.Host, cfg.Port)
}

// AMQPExchCfg represents the basic variables of any exchange
type AMQPExchCfg struct {
	ExchangeName string
	Durable      bool
}

// AMQPSubCfg represents the settings needed to create a new subscriber, including the queues and key bindings
type AMQPSubCfg struct {
	ExchangeName      string
	QueueID           uint64
	Keys              []string
	IsWorkQueue       bool
	HandleMessageFunc func(AMQPMessage) error
	Control           *RabbitControl
}

// QueueName generates the Queue
func (cfg AMQPSubCfg) QueueName() string {
	return RabbitWebsocketQueueName(cfg.QueueID)
}

// RabbitWebsocketQueueName returns the name of the Queue a websocket with the given ID would have
func RabbitWebsocketQueueName(queueID uint64) string {
	return fmt.Sprintf("%s-%d", hostname, queueID)
}

// RabbitProjectQueueName returns the name of the Queue a project with the given ID would have
func RabbitProjectQueueName(projectID int64) string {
	return fmt.Sprintf("Project-%d", projectID)
}

// AMQPPubCfg represents the settings needed to create a new publisher
type AMQPPubCfg struct {
	ExchangeName string
	Messages     chan AMQPMessage
	Control      *utils.Control
}

// NewPubConfig creates a new AMQPPubCfg, initialized
func NewPubConfig(exchangeName string) *AMQPPubCfg {
	cfg := AMQPPubCfg{
		ExchangeName: exchangeName,
		Messages:     make(chan AMQPMessage, 1),
	}
	cfg.Control = utils.NewControl()
	return &cfg
}

// AMQPMessage represents the information required to send a new message
type AMQPMessage struct {
	Headers     map[string]interface{}
	RoutingKey  string
	ContentType string
	Persistent  bool
	Message     []byte
}
