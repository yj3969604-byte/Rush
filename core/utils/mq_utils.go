package utils

import (
	"encoding/json"
	"errors"
	"github.com/wagslane/go-rabbitmq"
	"log"
)

var Publisher *rabbitmq.Publisher

func PublishMQ(mqMessage MQMessage) error {
	if Publisher == nil {
		log.Printf("publisher not exist")
		return errors.New("publisher not exist")
	}
	if mqMessage.MessageType == "" || mqMessage.Data == "" {
		log.Printf("mq data error.messageType=%s;data=%s", mqMessage.MessageType, mqMessage.Data)
		return errors.New("data error")
	}
	msgBytes, _ := json.Marshal(mqMessage)
	log.Printf("发送mq消息1:%s", string(msgBytes))
	return Publisher.Publish(
		msgBytes,
		[]string{GlobalConfig.RabbitMq.Routing},
		rabbitmq.WithPublishOptionsContentType("application/json"),
		rabbitmq.WithPublishOptionsExchange(GlobalConfig.RabbitMq.Exchange),
		rabbitmq.WithPublishOptionsExpiration("60000"),
	)
}

type MQMessage struct {
	MessageType string `json:"messageType"`
	Data        string `json:"data"`
	DataMore    string `json:"dataMore"`
}
