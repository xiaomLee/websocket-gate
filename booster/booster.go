package booster

import (
	"sync"
	"websocket-gate/logger"

	"github.com/gin-gonic/gin"

	"github.com/gorilla/websocket"
)

type Booster struct {
	upgrader websocket.Upgrader
	hubs     map[string]*Hub

	sync.WaitGroup
	sync.RWMutex
}

var booster *Booster

func InitBooster() {
	booster = NewBooster()
}

func GetBoosterInstance() *Booster {
	return booster
}

func NewBooster() *Booster {
	return &Booster{
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
		hubs: make(map[string]*Hub),
	}
}

func (b *Booster) getHub(topic string) *Hub {
	b.RLock()
	hub, ok := b.hubs[topic]
	b.RUnlock()

	if !ok {
		hub = newHub()
		b.Add(1)
		go func() {
			b.Done()
			hub.run()
		}()

		b.Lock()
		b.hubs[topic] = hub
		b.Unlock()
	}

	return hub
}

func (b *Booster) WsHandler(c *gin.Context) {
	c.Request.ParseForm()
	topic := c.Request.Form.Get("topic")

	conn, err := b.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logger.Error("WsAction", err, "upgrade err")
		c.String(200, err.Error())
		return
	}

	hub := b.getHub(topic)
	session := &Session{
		hub:  hub,
		conn: conn,
		send: make(chan []byte, 256),
	}

	session.hub.register <- session
	go session.writePump()
	go session.readPump()
}
