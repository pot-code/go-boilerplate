package infra

import (
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
	HandshakeTimeout: 3 * time.Second,
}

var (
	writeWait    = 10 * time.Second
	pongWait     = 30 * time.Second
	pingInterval = pongWait * 9 / 10
)

// WithHeartbeat wrap handler function with heartbeat probe
func WithHeartbeat(handler func(*websocket.Conn) error) echo.HandlerFunc {
	return func(c echo.Context) error {
		conn, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
		if err != nil {
			return c.String(http.StatusBadRequest, err.Error())
		}

		go heartbeatRoutine(conn)
		go processRoutine(conn, handler)
		return nil
	}
}

func heartbeatRoutine(conn *websocket.Conn) {
	ticker := time.NewTicker(pingInterval)
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})
	defer func() {
		ticker.Stop()
		conn.Close()
	}()
	for {
		select {
		case <-ticker.C:
			if err := conn.WriteControl(websocket.PingMessage, nil, time.Now().Add(writeWait)); err != nil {
				return
			}
		}
	}
}

func processRoutine(conn *websocket.Conn, handler func(*websocket.Conn) error) {
	defer conn.Close()
	for {
		if err := handler(conn); err != nil {
			break
		}
	}
}
