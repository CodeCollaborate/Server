package datahandling

import (
	"github.com/CodeCollaborate/Server/modules/datahandling/messages"
	"github.com/CodeCollaborate/Server/modules/dbfs"
	"github.com/CodeCollaborate/Server/modules/rabbitmq"
	"github.com/CodeCollaborate/Server/utils"
)

/**
 * Data Handling logic for the CodeCollaborate Server.
 */

// DataHandler handles the json data received from the WebSocket connection.
type DataHandler struct {
	MessageChan chan<- rabbitmq.AMQPMessage
	Db          dbfs.DBFS
}

// Handle takes the MessageType and message in byte-array form,
// processing the data, and updating DBFS/RabbitMQ as needed.
func (dh DataHandler) Handle(message []byte, origin string, ack func() error) error {
	utils.LogDebug("Received Message", utils.LogFields{
		"Message": string(message),
	})

	req, err := createAbstractRequest(message)
	if err != nil {
		utils.LogError("Failed to parse json", err, nil) // Do not log request since passwords may be sent
		ack()
		return err
	}

	// automatically determines if the request is authenticated or not
	fullRequest, err := getFullRequest(req)

	var closures []dhClosure
	if err != nil {
		utils.LogDebug("Authentication Error", utils.LogFields{
			"Resource": req.Resource,
			"Method":   req.Method,
			"SenderID": req.SenderID,
			"Error":    err,
		})
		ack()
		closures = []dhClosure{toSenderClosure{msg: messages.NewEmptyResponse(messages.StatusUnauthorized, req.Tag)}}
	} else {
		closures, err = fullRequest.process(dh.Db, ack)
		if err != nil {
			utils.LogError("Failed to process request", err, utils.LogFields{
				"Resource": req.Resource,
				"Method":   req.Method,
			})
		}
	}

	for _, closure := range closures {
		err := closure.call(dh, origin)
		if err != nil {
			utils.LogError("Failed to complete continuation", err, utils.LogFields{
				"Resource": req.Resource,
				"Method":   req.Method,
			})
		}
	}

	return err
}
