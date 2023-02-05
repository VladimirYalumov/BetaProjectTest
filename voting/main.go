package main

import (
	"BetaProjectTest/log_consumer/logger"
	"BetaProjectTest/log_consumer/rabbit"
	"BetaProjectTest/voting/orm"
	"BetaProjectTest/voting/services"
	"bytes"
	"encoding/json"
	amqp "github.com/rabbitmq/amqp091-go"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
	"time"
)

var conn *amqp.Connection
var ch *amqp.Channel

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
	var client rabbit.Client
	err := client.SetConnection("guest", "guest", "127.0.0.1:5673")
	if err != nil {
		panic(err)
	}
	logInstance := logger.GetInstance()
	logInstance.InitPusher(client)
	orm.InitDB("127.0.0.1", "user", "qwerty", "deewave", "5432")
}

func main() {
	defer func() {
		if r := recover(); r != nil {
			datas := make(map[string]interface{}) // it is only example
			logger.GetInstance().DefinePid("", "default", "")
			datas["msg"] = "critical error"
			datas["body"] = r
			_ = logger.GetInstance().Info(datas)
		}
	}()
	ch, globalConnectionError = conn.Channel()
	voteTypes := NewCounters()
	countersOptionsOld := NewOldCounters()
	counters := NewCounters()

	countsGrouped := orm.FindAllVotes()
	for _, v := range countsGrouped {
		key := v.Vid
		value, ok := voteTypes.Load(key)
		if ok {
			voteTypes.Store(key, v.Count+value)
		} else {
			voteTypes.Store(key, v.Count)
		}
		key = services.MakeKey(v.Vid, v.Oid)
		value, ok = counters.Load(key)
		if ok {
			counters.Store(key, v.Count+value)
		} else {
			counters.Store(key, v.Count)
		}
	}

	blockCh := make(chan map[string]interface{})

	for _, v := range countsGrouped {
		votesInOption, _ := counters.Load(services.MakeKey(v.Vid, v.Oid))
		option, _ := voteTypes.Load(v.Vid)
		countersOptionsOld.Store(services.MakeKey(v.Vid, v.Oid), float64(votesInOption)/float64(option)*100)
	}

	var msgs <-chan amqp.Delivery
	msgs, globalConnectionError = ch.Consume(
		services.VotingRoute, // queue
		"",                   // consumer
		true,                 // auto-ack
		false,                // exclusive
		false,                // no-local
		false,                // no-wait
		nil,                  // args
	)

	if globalConnectionError != nil {
		panic(globalConnectionError)
	}
	if globalConnectionError != nil {
		panic(globalConnectionError)
	}

	var forever chan struct{}

	i := 0
	for {
		// не более 2 запросов в секунду
		if i >= 2 {
			break
		}
		go func() {
			for {
				t := <-blockCh
				datas := make(map[string]interface{}) // it is only example
				datas["msg"] = "send request to gamma"
				datas["body"] = t
				_ = logger.GetInstance().Info(datas)
				postBody, _ := json.Marshal(t)
				resp, err := http.Post(
					"http://127.0.0.1:8090/voting-stats",
					"application/json",
					bytes.NewBuffer(postBody),
				)
				_ = resp.Body.Close()

				if err != nil {
					datas["msg"] = "response from gamma error"
					datas["body"] = err.Error()
					_ = logger.GetInstance().Error(datas)
				} else {
					body, _ := ioutil.ReadAll(resp.Body)
					datas["msg"] = "response from gamma success"
					datas["body"] = string(body)
					_ = logger.GetInstance().Info(datas)
				}

				_ = resp.Body.Close()
				time.Sleep(1 * time.Second)
			}
		}()
		i++
	}

	go func() {
		defer func() {
			if r := recover(); r != nil {
				datas := make(map[string]interface{}) // it is only example
				logger.GetInstance().DefinePid("", "default", "")
				datas["msg"] = "critical error"
				datas["body"] = r
				_ = logger.GetInstance().Info(datas)
			}
		}()
		for d := range msgs {
			logger.GetInstance().DefinePid("", services.VotingRoute, d.MessageId)
			key, voteTypesKey := services.HandleRequest(d)
			if key != "" {
				o, ok := voteTypes.Load(voteTypesKey)
				first := false
				if ok {
					voteTypes.Inc(voteTypesKey)
					o++
				} else {
					voteTypes.Store(voteTypesKey, 1)
					o = 1
					first = true
				}
				v, ok := counters.Load(key)
				if ok {
					counters.Inc(key)
					v++
				} else {
					counters.Store(key, 1)
					v = 1
					first = true
				}

				value, ok := countersOptionsOld.Load(key)
				counterOptionsNew := float64(v) / float64(o) * 100
				if (ok && (counterOptionsNew-value) > 1) || first {
					Key := strings.Split(key, "_")
					countersOptionsOld.Store(key, counterOptionsNew)
					temporaryMap := countersOptionsOld.m
					blockCh <- map[string]interface{}{
						"votingId": voteTypesKey,
						"results":  getResults(Key[0], temporaryMap, counters),
					}
				}
			}
		}
	}()

	<-forever
}

func getResults(currentVouteType string, countersOptionsOld map[string]float64, counters *Counters) (results []map[string]interface{}) {
	for currentKey, _ := range countersOptionsOld {
		localKey := strings.Split(currentKey, "_")
		if localKey[0] == currentVouteType {
			count, _ := counters.Load(currentKey)
			results = append(results, map[string]interface{}{
				"optionId": localKey[1],
				"count":    count,
			})
		}
	}
	return
}

type Counters struct {
	mx sync.Mutex
	m  map[string]int
}

type OldCounters struct {
	mx sync.Mutex
	m  map[string]float64
}

func NewCounters() *Counters {
	return &Counters{
		m: make(map[string]int),
	}
}

func (c *Counters) Load(key string) (int, bool) {
	c.mx.Lock()
	defer c.mx.Unlock()
	val, ok := c.m[key]
	return val, ok
}

func (c *Counters) Store(key string, value int) {
	c.mx.Lock()
	defer c.mx.Unlock()
	c.m[key] = value
}

func (c *Counters) Inc(key string) {
	c.mx.Lock()
	defer c.mx.Unlock()
	c.m[key]++
}

func NewOldCounters() *OldCounters {
	return &OldCounters{
		m: make(map[string]float64),
	}
}

func (c *OldCounters) Load(key string) (float64, bool) {
	c.mx.Lock()
	defer c.mx.Unlock()
	val, ok := c.m[key]
	return val, ok
}

func (c *OldCounters) Store(key string, value float64) {
	c.mx.Lock()
	defer c.mx.Unlock()
	c.m[key] = value
}
