package handlers

import (
	"context"
	"net/http"

	"github.com/Invan2/invan_websocket_service/pkg/logger"
	"github.com/gorilla/websocket"
)

type handler struct {
	upgrader *websocket.Upgrader
	log      logger.Logger
	ctx      context.Context
}

type Handler interface {
	Main(w http.ResponseWriter, r *http.Request)
}

func NewWebSocketHandler(upgrader *websocket.Upgrader, log logger.Logger, ctx context.Context) Handler {

	return &handler{
		upgrader: upgrader,
		log:      log,
		ctx:      ctx,
	}
}

func (h *handler) Main(w http.ResponseWriter, r *http.Request) {

	c, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.log.Error("main handler", logger.Error(err))
		return
	}

	defer c.Close()

	for {
		select {
		case <-h.ctx.Done():
			return
		default:
			mt, messages, err := c.ReadMessage()
			if err != nil {
				h.log.Error("error while read message")
				return
			}

			if string(messages) == "PING" {
				c.WriteMessage(mt, []byte("PONG"))
			}

		}

	}

}
