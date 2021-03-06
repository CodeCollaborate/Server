package datahandling

import (
	"encoding/json"
	"errors"

	"github.com/CodeCollaborate/Server/modules/datahandling/messages"
	"github.com/CodeCollaborate/Server/modules/rabbitmq"
	"github.com/CodeCollaborate/Server/utils"
)

type dhClosure interface {
	call(dh DataHandler) error
}

type toSenderClosure struct {
	msg *messages.ServerMessageWrapper
}

// toSenderClosure.call is the function that will forward a server message back to the client
func (cont toSenderClosure) call(dh DataHandler) error {
	msgJSON, err := json.Marshal(cont.msg)
	if err != nil {
		return err
	}

	msg := rabbitmq.AMQPMessage{
		Headers: map[string]interface{}{
			"Origin":      rabbitmq.RabbitWebsocketQueueName(dh.WebsocketID),
			"MessageType": cont.msg.Type,
		},
		RoutingKey:  rabbitmq.RabbitWebsocketQueueName(dh.WebsocketID),
		ContentType: rabbitmq.ContentTypeMsg,
		Persistent:  false,
		Message:     msgJSON,
	}

	select {
	case dh.MessageChan <- msg:
	default:
		utils.LogError("AMQP Publisher message queue full; failed to add new message", errors.New("Channel buffer full"), utils.LogFields{
			"AMQP Message": msg,
		})
		return errors.New("Channel buffer full")
	}
	return nil
}

type toRabbitChannelClosure struct {
	msg *messages.ServerMessageWrapper
	key string
}

// toRabbitChannelClosure.call is the function that will forward a server message to a channel based on the given routing key
func (cont toRabbitChannelClosure) call(dh DataHandler) error {
	msgJSON, err := json.Marshal(cont.msg)
	if err != nil {
		return err
	}

	msg := rabbitmq.AMQPMessage{
		Headers: map[string]interface{}{
			"Origin":      rabbitmq.RabbitWebsocketQueueName(dh.WebsocketID),
			"MessageType": cont.msg.Type,
		},
		RoutingKey:  cont.key,
		ContentType: rabbitmq.ContentTypeMsg,
		Persistent:  false,
		Message:     msgJSON,
	}

	select {
	case dh.MessageChan <- msg:
	default:
		utils.LogError("AMQP Publisher message queue full; failed to add new message", errors.New("Channel buffer full"), utils.LogFields{
			"AMQP Message": msg,
		})
		return errors.New("Channel buffer full")
	}

	return nil
}

type rabbitCommandClosure struct {
	Command string
	Tag     int64
	Key     string
	Data    interface{}
}

// toRabbitChannelClosure.call is the function that will forward a server message to a channel based on the given routing key
func (cont rabbitCommandClosure) call(dh DataHandler) error {
	msgJSON, err := json.Marshal(cont)
	if err != nil {
		return err
	}

	if cont.Key == "" {
		cont.Key = rabbitmq.RabbitWebsocketQueueName(dh.WebsocketID)
	}

	msg := rabbitmq.AMQPMessage{
		Headers: map[string]interface{}{
			"Origin": rabbitmq.RabbitWebsocketQueueName(dh.WebsocketID),
		},
		RoutingKey:  cont.Key,
		ContentType: rabbitmq.ContentTypeCmd,
		Persistent:  false,
		Message:     msgJSON,
	}

	select {
	case dh.MessageChan <- msg:
	default:
		utils.LogError("AMQP Publisher message queue full; failed to add new message", errors.New("Channel buffer full"), utils.LogFields{
			"AMQP Message": msg,
		})
		return errors.New("Channel buffer full")
	}

	return nil
}
