package websocket

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Invan2/invan_websocket_service/config"
	"github.com/Invan2/invan_websocket_service/pkg/logger"
	"github.com/Invan2/invan_websocket_service/websocket/handlers"
	"github.com/gorilla/websocket"
)

type websocketServer struct {
	log logger.Logger
	cfg config.Config
}

type WebsocketServer interface {
	Run(ctx context.Context) error
}

func NewWebSocketServer(log logger.Logger, cfg config.Config) (WebsocketServer, error) {
	return &websocketServer{
		log: log,
		cfg: cfg,
	}, nil
}

func (w *websocketServer) Run(ctx context.Context) error {
	upgrader := websocket.Upgrader{}

	upgrader.EnableCompression = true
	upgrader.CheckOrigin = func(r *http.Request) bool {
		return true
	}

	handler := handlers.NewWebSocketHandler(&upgrader, w.log, ctx)

	http.HandleFunc("/", handler.Main)

	return http.ListenAndServe(fmt.Sprintf("%s:%d", w.cfg.HTTPHost, w.cfg.HTTPPort), nil)
}
