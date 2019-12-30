package booster

// Hub maintains the set of active sessions and broadcasts messages to the
// sessions.
type Hub struct {
	// Registered sessions.
	sessions map[*Session]bool

	// Inbound messages from the sessions.
	broadcast chan []byte

	// Register requests from the sessions.
	register chan *Session

	// Unregister requests from sessions.
	unregister chan *Session

	// exit
	breakLogic chan bool
	exited     chan bool
}

func newHub() *Hub {
	return &Hub{
		broadcast:  make(chan []byte),
		register:   make(chan *Session),
		unregister: make(chan *Session),
		sessions:   make(map[*Session]bool),
		breakLogic: make(chan bool),
		exited:     make(chan bool),
	}
}

func (h *Hub) run() {
	for {
		select {
		case s := <-h.register:
			h.sessions[s] = true
		case s := <-h.unregister:
			if _, ok := h.sessions[s]; ok {
				delete(h.sessions, s)
				close(s.send)
			}
		case message := <-h.broadcast:
			for client := range h.sessions {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.sessions, client)
				}
			}

		case <-h.breakLogic:
			// todo stop sessions
			for s := range h.sessions {
				s.stop()
			}
			goto EXIT
		}
	}
EXIT:
	h.exited <- true
}

func (h *Hub) stop() {
	close(h.breakLogic)
	<-h.exited
}
