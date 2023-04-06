package main

import (
	"context"

	"github.com/Invan2/invan_websocket_service/config"
	"github.com/Invan2/invan_websocket_service/models"
	"github.com/Invan2/invan_websocket_service/pkg/kafka"
	"github.com/Invan2/invan_websocket_service/pkg/logger"
	"github.com/Invan2/invan_websocket_service/websocket"
	v2 "github.com/cloudevents/sdk-go/v2"
)

func main() {

	cfg := config.Load()

	log := logger.New(cfg.LogLevel, cfg.ServiceName)

	ctx, cancel := context.WithCancel(context.Background())

	defer cancel()

	k, err := kafka.NewKafka(ctx, cfg, log)
	if err != nil {
		log.Error("kafka", logger.Error(err))
	}

	k.AddConsumer("hello", func(ctx context.Context, e v2.Event) models.Response {
		return models.Response{}
	})

	ws, err := websocket.NewWebSocketServer(log, cfg)
	if err != nil {
		log.Error("error while creating web socket service", logger.Error(err))
		return
	}

	ws.Run(ctx)

}
