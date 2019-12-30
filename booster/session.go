package booster

import (
	"bytes"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512

	//
	msgPing = "ping"
	msgPong = "pong"
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

// Session is a middleman between the websocket connection and the booster.
type Session struct {
	// The websocket connection.
	conn *websocket.Conn

	// Buffered channel of outbound messages.
	send chan []byte

	sync.WaitGroup
}

func newSession(conn *websocket.Conn) *Session {
	return &Session{
		conn: conn,
		send: make(chan []byte, 256),
	}
}

func (s *Session) run() {
	// write
	go func() {
		s.Add(1)
		s.writePump()
		s.Done()
	}()
	// read
	go func() {
		s.Add(1)
		s.readPump()
		s.Done()
	}()
}

func (s *Session) stop() {
	s.conn.WriteMessage(websocket.CloseMessage, []byte{})
	close(s.send)
}

// readPump pumps messages from the websocket connection to the booster.
// The application runs readPump in a per-connection goroutine. The application
// ensures that there is at most one reader on a connection by executing all
// reads from this goroutine.
func (s *Session) readPump() {
	defer s.conn.Close()

	s.conn.SetReadLimit(maxMessageSize)
	s.conn.SetReadDeadline(time.Now().Add(pongWait))
	s.conn.SetPongHandler(func(string) error { s.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		t, message, err := s.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}

		if t == websocket.CloseMessage {
			break
		}

		message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))
		if string(message) == msgPing {
			s.send <- []byte(msgPong)
			continue
		}

		messageHandler(s, message)
	}
}

// writePump pumps messages from the booster to the websocket connection.
// A goroutine running writePump is started for each connection. The
// application ensures that there is at most one writer to a connection by
// executing all writes from this goroutine.
func (s *Session) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		s.conn.Close()
	}()

	for {
		select {
		case message, ok := <-s.send:
			s.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The booster closed the channel.
				s.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := s.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued chat messages to the current websocket message.
			n := len(s.send)
			for i := 0; i < n; i++ {
				w.Write(newline)
				w.Write(<-s.send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			if err := s.ping(); err != nil {
				return
			}
		}
	}
}

func (s *Session) ping() error {
	s.conn.SetWriteDeadline(time.Now().Add(writeWait))
	if err := s.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
		return err
	}
	return nil
}
