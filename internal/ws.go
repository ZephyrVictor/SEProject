package internal

import (
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var (
	upgrader = websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
	hubs     = make(map[string]map[*websocket.Conn]bool)
	hubsMu   sync.Mutex
)

func HandleWs(c *gin.Context) {
	email := c.GetString("email")
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	hubsMu.Lock()
	if hubs[email] == nil {
		hubs[email] = make(map[*websocket.Conn]bool)
	}

	hubs[email][conn] = true
	hubsMu.Unlock()

	for {
		if _, _, err := conn.NextReader(); err != nil {
			hubsMu.Lock()
			delete(hubs[email], conn)
			hubsMu.Unlock()
			conn.Close()
			break
		}
	}
}

// 推送消息
func NotifyUser(email string, msg interface{}) {
	hubsMu.Lock()
	conns := hubs[email]
	hubsMu.Unlock()
	for c := range conns {
		c.WriteJSON(msg)
	}
}
