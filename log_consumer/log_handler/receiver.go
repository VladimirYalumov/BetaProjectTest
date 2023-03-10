package log_handler

import (
	"BetaProjectTest/log_consumer/mongo"
	"BetaProjectTest/log_consumer/rabbit"
)

func Distribute(message []byte) error {
	msg := rabbit.MessageBody{}
	err := msg.Decode(message)
	if err != nil {
		return err
	}
	return mongo.InsertValue(msg.Body, msg.Pid, msg.CreatedTime, msg.Action, msg.Level)
}
