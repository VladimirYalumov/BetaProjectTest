package services

import (
	"BetaProjectTest/log_consumer/logger"
	"BetaProjectTest/request"
	"BetaProjectTest/voting/orm"
	"encoding/json"
	"github.com/rabbitmq/amqp091-go"
	"strconv"
)

const VotingRoute = "voting"

type RabbitRequest struct {
	Body request.VotingRequest `json:"Body"`
}

func HandleRequest(msg amqp091.Delivery) (string, string) {
	datas := make(map[string]interface{})
	datas["body"] = string(msg.Body)
	var req request.VotingRequest
	err := json.Unmarshal(msg.Body, &req)
	if err != nil {
		datas["msg"] = "Invalid params: " + err.Error()
		_ = logger.GetInstance().Error(datas)
		return "", ""
	}

	dbErr := saveVote(req)
	if dbErr != nil {
		datas["msg"] = "Insert to db error: " + dbErr.Error()
		_ = logger.GetInstance().Error(datas)
		return "", ""
	}

	datas["msg"] = "Success"
	_ = logger.GetInstance().Info(datas)
	return MakeKey(strconv.Itoa(req.VotingId), strconv.Itoa(req.OptionId)), strconv.Itoa(req.VotingId)
}

func saveVote(req request.VotingRequest) error {
	vote := orm.Vote{
		VoteId:   req.VoteId,
		VotingId: req.VotingId,
		OptionId: req.OptionId,
	}
	return vote.Create()
}

func MakeKey(vid string, oid string) string {
	return vid + "_" + oid
}
