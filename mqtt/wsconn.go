package mqtt

import (
	"net/http"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

var (
	upGrader = &websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		Subprotocols:    []string{"mqtt", "mqtts"},
	}
)

func httpWorker(w http.ResponseWriter, r *http.Request) {
	conn, err := upGrader.Upgrade(w, r, make(http.Header))
	if err != nil {
		log.Error("establish ws connection fail", zap.Error(err))
		return
	}
	defer conn.Close()
}
