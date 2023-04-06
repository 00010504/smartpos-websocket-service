package kafka

import (
	"context"
	"errors"

	"github.com/Invan2/invan_websocket_service/models"
	"github.com/Invan2/invan_websocket_service/pkg/helper"
	"github.com/Invan2/invan_websocket_service/pkg/logger"
	"github.com/Shopify/sarama"
	cloudevents "github.com/cloudevents/sdk-go/v2"
)

type HandlerFunc func(context.Context, cloudevents.Event) models.Response

type Consumer struct {
	consumerName string
	topic        string
	handler      HandlerFunc
}

func (kafka *Kafka) AddConsumer(topic string, handler HandlerFunc) {
	if kafka.consumers[topic] != nil {
		panic(errors.New("consumer with the same name already exists: " + topic))
	}

	kafka.consumers[topic] = &Consumer{
		consumerName: topic,
		topic:        topic,
		handler:      handler,
	}
}

func (c *Kafka) Setup(_ sarama.ConsumerGroupSession) error {
	close(c.ready)
	return nil
}

func (c *Kafka) Cleanup(_ sarama.ConsumerGroupSession) error {
	return nil
}

func (c *Kafka) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	consumer := c.consumers[claim.Topic()]
	for message := range claim.Messages() {

		event := helper.MessageToEvent(message)

		session.MarkMessage(message, "")
		resp := consumer.handler(c.ctx, event)

		if resp.Topic == "" {
			return nil
		}

		err := event.SetData(cloudevents.ApplicationJSON, resp)
		if err != nil {
			c.log.Error("Failed to set data", logger.Any("error:", err))
		}

		err = c.Push(resp.Topic, event)
		if err != nil {
			c.log.Error("Failed to push", logger.Any("error:", err))
		}

	}
	return nil
}
