package booster

import (
	"net/http"
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

func CloseBooster() {
	for _, h := range booster.hubs {
		h.stop()
	}
	booster.Wait()
}

func NewBooster() *Booster {
	return &Booster{
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin:     checkOrigin,
		},
		hubs: make(map[string]*Hub),
	}
}

func checkOrigin(r *http.Request) bool {
	// todo

	return true
}

func (b *Booster) getHub(topic string) *Hub {
	b.RLock()
	hub, ok := b.hubs[topic]
	b.RUnlock()

	if !ok {
		hub = newHub()
		go func() {
			b.Add(1)
			defer b.Done()
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
		logger.Error("WsHandler", err, "upgrade err")
		c.String(200, err.Error())
		return
	}

	hub := b.getHub(topic)
	session := newSession(conn)

	connectHandler(session)
	hub.register <- session

	// run write and read
	session.run()
	// wait for session close
	session.Wait()

	hub.unregister <- session
	disConnectHandler(session)
}

func (b *Booster) PushMessage() {

}

func disConnectHandler(s *Session) {
	println("disconnect")
}

func messageHandler(s *Session, message []byte) {
	println(string(message))
	// todo 判断msg类型 是否broadcast
	// broadcast
	//s.hub.broadcast <- message
}

func connectHandler(s *Session) {
	println("connect")
}

func errHandler(s *Session, err error) {
	println("error:" + err.Error())
}
