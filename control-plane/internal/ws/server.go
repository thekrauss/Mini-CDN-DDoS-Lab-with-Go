package ws

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{}

func NewWSServer(hub *Hub) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("Erreur upgrade websocket: %v", err)
			return
		}
		hub.Register(conn)

		go func() {
			defer hub.Unregister(conn)
			for {
				if _, _, err := conn.NextReader(); err != nil {
					break
				}
			}
		}()
	})
}
