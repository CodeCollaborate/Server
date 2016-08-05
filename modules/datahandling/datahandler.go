package datahandling

import (
	"fmt"

	"encoding/json"

	"strconv"

	"time"

	"github.com/CodeCollaborate/Server/modules/dbfs"
	"github.com/CodeCollaborate/Server/modules/rabbitmq"
	"github.com/CodeCollaborate/Server/utils"
)

/**
 * Data Handling logic for the CodeCollaborate Server.
 */

// DataHandler handles the json data received from the WebSocket connection.
type DataHandler struct {
	MessageChan      chan<- rabbitmq.AMQPMessage
	SubscriptionChan chan<- rabbitmq.Subscription
	WebsocketID      uint64
	Db               dbfs.DBFS
}

// Handle takes the WebSocket Id, MessageType and message in byte-array form,
// processing the data, and updating DB/FS/RabbitMQ as needed.
func (dh DataHandler) Handle(messageType int, message []byte) error {
	fmt.Printf("Handling Message: %s\n", message)

	req, err := createAbstractRequest(message)
	if err != nil {
		utils.LogOnError(err, "Failed to parse json")
		return err
	}

	// automatically determines if the request is authenticated or not
	fullRequest, err := getFullRequest(req)

	var closures []dhClosure

	if err != nil {
		res := new(serverMessageWrapper)
		res.Timestamp = time.Now().UnixNano()
		res.Type = "Responce"
		if err == ErrAuthenticationFailed {
			utils.LogOnError(err, "User not logged in")
			res.ServerMessage = response{
				Status: unauthorized,
				Tag:    req.Tag,
				Data:   struct{}{}}
		} else {
			utils.LogOnError(err, "Failed to construct full request")
			res.ServerMessage = response{
				Status: unimplemented,
				Tag:    req.Tag,
				Data:   struct{}{}}
		}
		closures = []dhClosure{toSenderClosure{msg: res}}
	} else {
		closures, err = fullRequest.process(dh.Db)
		if err != nil {
			utils.LogOnError(err, "Failed to handle process request")
		}
	}

	for _, closure := range closures {
		erro := closure.call(dh)
		if erro != nil {
			utils.LogOnError(erro, "Failed to complete continuation")
		}
	}

	return err
}

func authenticate(abs abstractRequest) bool {
	fmt.Println("AUTHENTICATION IS NOT IMPLEMENTED YET")
	// TODO (non-immediate/required): implement user authentication
	return true
}

// SendToSender is the function that will forward a server message back to the client
func (dh DataHandler) sendToSender(msg *serverMessageWrapper) error {
	msgJSON, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	dh.MessageChan <- rabbitmq.AMQPMessage{
		Headers:     make(map[string]interface{}),
		RoutingKey:  strconv.FormatUint(dh.WebsocketID, 10),
		ContentType: msg.Type,
		Persistent:  false,
		Message:     msgJSON,
	}
	return nil
}

// SendToChannel is the function that will forward a server message to a channel based on the given routing key
func (dh DataHandler) sendToChannel(msg *serverMessageWrapper) error {
	msgJSON, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	dh.MessageChan <- rabbitmq.AMQPMessage{
		Headers:     make(map[string]interface{}),
		RoutingKey:  msg.RoutingKey,
		ContentType: msg.Type,
		Persistent:  false,
		Message:     msgJSON,
	}
	return nil
}
