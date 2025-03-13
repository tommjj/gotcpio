package gotcpio

import "sync"

type Room struct {
	conns map[Emiter]struct{}
	mu    sync.RWMutex
}

// Add add a emiter to the room.
func (r *Room) Add(e Emiter) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.conns[e] = struct{}{}
}

// Remove removes a emiter from the room.
func (r *Room) Remove(e Emiter) {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.conns, e)
}

// Emit sends an event with the given body to all emitters in the room.
func (r *Room) Emit(event string, body []byte) error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for conn := range r.conns {
		conn.Emit(event, body)
	}
	return nil
}

// Size returns the number of emitters in the room.
func (r *Room) Size() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return len(r.conns)
}
