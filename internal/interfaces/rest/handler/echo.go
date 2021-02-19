package handler

import (
	"github.com/gorilla/websocket"
)

// HandleEcho echo message back handler
func HandleEcho(conn *websocket.Conn) error {
	_, message, err := conn.ReadMessage()
	if err != nil {
		return err
	}
	if err = conn.WriteMessage(websocket.TextMessage, append([]byte("Echo: "), message...)); err != nil {
		return err
	}
	return nil
}
