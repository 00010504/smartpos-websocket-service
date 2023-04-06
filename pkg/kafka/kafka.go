package kafka

import (
	"context"
	"fmt"
	"sync"
	"time"

	// "go_boilerplate/pkg/logger"

	"github.com/Invan2/invan_websocket_service/config"
	"github.com/Invan2/invan_websocket_service/pkg/logger"
	"github.com/Shopify/sarama"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/google/uuid"
)

type Kafka struct {
	ctx           context.Context
	log           logger.Logger
	cfg           config.Config
	consumers     map[string]*Consumer
	publishers    map[string]*Publisher
	saramaConfig  *sarama.Config
	consumerGroup sarama.ConsumerGroup
	ready         chan struct{}
	wg            *sync.WaitGroup
}

type KafkaI interface {
	RunConsumers()
	AddConsumer(topic string, handler HandlerFunc)
	Push(topic string, e cloudevents.Event) error
	AddPublisher(topic string)
	Shutdown() error
}

func NewKafka(ctx context.Context, cfg config.Config, log logger.Logger) (KafkaI, error) {
	saramaConfig := sarama.NewConfig()
	saramaConfig.Version = sarama.V2_0_0_0
	saramaConfig.Consumer.Offsets.Initial = sarama.OffsetOldest
	saramaConfig.Consumer.Group.Heartbeat.Interval = time.Second * 30
	saramaConfig.Consumer.Group.Session.Timeout = time.Second * 90
	saramaConfig.Consumer.Group.Rebalance.Timeout = time.Second * 90 * 3
	saramaConfig.Producer.MaxMessageBytes = 1024 * 1024 * 40
	saramaConfig.Consumer.MaxProcessingTime = time.Second * 60

	consumerGroup, err := sarama.NewConsumerGroup([]string{cfg.KafkaUrl}, config.ConsumerGroupID, saramaConfig)
	if err != nil {
		return nil, err
	}

	kafka := &Kafka{
		ctx:           ctx,
		log:           log,
		cfg:           cfg,
		consumers:     make(map[string]*Consumer),
		publishers:    make(map[string]*Publisher),
		saramaConfig:  saramaConfig,
		ready:         make(chan struct{}),
		wg:            &sync.WaitGroup{},
		consumerGroup: consumerGroup,
	}

	return kafka, nil
}

// RunConsumers ...
func (r *Kafka) RunConsumers() {
	topics := []string{}

	for _, consumer := range r.consumers {
		topics = append(topics, consumer.topic)
		fmt.Println("Key:", consumer.topic, "=>", "consumer:", consumer)
	}
	r.log.Info("topics:", logger.Any("topics:", topics))

	r.wg.Add(1)
	go func() {
		defer r.wg.Done()
		for {
			if err := r.consumerGroup.Consume(r.ctx, topics, r); err != nil {
				r.log.Error("error while consuming", logger.Error(err))
			}
			if r.ctx.Err() != nil {
				return
			}
			r.ready = make(chan struct{})
		}
	}()

	<-r.ready
	r.log.Warn("consumer group started")
}

func CreateEvent(t, s string, v interface{}) (cloudevents.Event, error) {
	event := cloudevents.NewEvent()
	id, err := uuid.NewRandom()
	if err != nil {
		return event, err
	}
	event.SetType(t)
	event.SetSource(s)
	event.SetID(id.String())
	err = event.SetData(cloudevents.ApplicationJSON, v)
	return event, err
}

func (r *Kafka) Shutdown() error {
	r.log.Warn("shutting down pub-sub server")
	select {
	case <-r.ctx.Done():
		r.log.Warn("terminating: context cancelled")
	default:
	}
	r.wg.Wait()
	r.consumerGroup.Close()

	for _, publisher := range r.publishers {
		if err := publisher.sender.Close(context.Background()); err != nil {
			r.log.Error("could not close sender", logger.Any("topic", publisher.topic), logger.Error(err))
		}
	}

	return nil
}
