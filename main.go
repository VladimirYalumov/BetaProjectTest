package main

import (
	"BetaProjectTest/log_consumer/logger"
	"BetaProjectTest/log_consumer/rabbit"
	"BetaProjectTest/response"
	"BetaProjectTest/voting/services"
	"encoding/json"
	amqp "github.com/rabbitmq/amqp091-go"
	"io/ioutil"
	"log"
	"net/http"
)

var conn *amqp.Connection
var ch *amqp.Channel
var votingQueue amqp.Queue

var globalConnectionError error

func init() {
	conn, globalConnectionError = amqp.Dial("amqp://guest:guest@localhost:5673/")

	if globalConnectionError != nil {
		panic(globalConnectionError)
	}

	ch, globalConnectionError = conn.Channel()

	if globalConnectionError != nil {
		panic(globalConnectionError)
	}

	votingQueue, globalConnectionError = ch.QueueDeclare(
		services.VotingRoute, // name
		false,                // durable
		false,                // delete when unused
		false,                // exclusive
		false,                // no-wait
		nil,                  // arguments
	)

	if globalConnectionError != nil {
		panic(globalConnectionError)
	}

	// инициализируем логгер
	var client rabbit.Client
	err := client.SetConnection("guest", "guest", "127.0.0.1:5673")
	if err != nil {
		panic(err)
	}
	logInstance := logger.GetInstance()
	logInstance.InitPusher(client)
}

func main() {
	http.HandleFunc("/"+services.VotingRoute, func(w http.ResponseWriter, r *http.Request) {
		logger.GetInstance().DefinePid(r.Host, services.VotingRoute, "")
		bodyString, _ := ioutil.ReadAll(r.Body)
		logRequest(bodyString, r.Host)
		pushRequest(bodyString)
		w.WriteHeader(200)
		_ = json.NewEncoder(w).Encode(&response.VotingResponse{Result: "OK"})
		return
	})

	log.Fatal(http.ListenAndServe(":8089", nil))
}

func logRequest(r []byte, host string) {
	datas := make(map[string]interface{}) // it is only example
	datas["msg"] = "start Action: voting"
	datas["user_ip"] = host
	datas["body"] = r
	_ = logger.GetInstance().Info(datas)
}

func pushRequest(req []byte) {
	err := ch.Publish(
		"",               // exchange
		votingQueue.Name, // routing key
		false,            // mandatory
		false,            // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        req,
			MessageId:   logger.GetInstance().GetPid(),
		})
	if err != nil {
		datas := make(map[string]interface{})
		datas["error"] = err.Error()
		_ = logger.GetInstance().Fatal(datas)
	}
}
