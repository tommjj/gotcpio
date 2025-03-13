package gotcpio

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"sync"

	"github.com/google/uuid"
)

type Data []byte

// EventHandler is a function that handles an event.
type EventHandler func(Data)

type Conn struct {
	id   uuid.UUID
	Conn net.Conn

	// Context is the context of the connection.
	context context.Context
	cal     context.CancelFunc

	events   map[string]EventHandler
	eventMux sync.RWMutex

	// store is a map of key-value pairs.
	store sync.Map
}

func (c *Conn) ID() uuid.UUID {
	return c.id
}

// NewConn creates a new Conn.
func NewConn(ctx context.Context, conn net.Conn) *Conn {
	connCtx, cal := context.WithCancel(ctx)

	return &Conn{
		Conn: conn,

		context: connCtx,
		cal:     cal,

		events: make(map[string]EventHandler),

		store: sync.Map{},

		id: uuid.New(),
	}
}

// ********************
// ** Event Handling **
// ********************

// ON registers an event handler for the given event.
func (c *Conn) On(event string, handler EventHandler) {
	c.eventMux.Lock()
	defer c.eventMux.Unlock()

	c.events[event] = handler
}

// Off removes the event handler for the given event.
func (c *Conn) Off(event string) {
	c.eventMux.Lock()
	defer c.eventMux.Unlock()

	delete(c.events, event)
}

// EMIT sends an event with the given body to the client.
func (c *Conn) Emit(event string, body []byte) error {
	_, err := c.Conn.Write([]byte(event + " " + string(body) + "\x00"))
	return err
}

// handleEvent handles an event with the given body.
// It calls the event handler for the event if it exists.
// If the event handler panics, it recovers and prints the error.
func (c *Conn) handleEvent(event string, body []byte) {
	c.eventMux.RLock()
	handler, ok := c.events[event]
	c.eventMux.RUnlock()

	if !ok {
		return
	}

	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered from panic in event handler:", event, r)
			return
		}
	}()

	handler(body)
}

// *************************
// ********* Store *********
// *************************

// Set sets a key-value pair in the connection's store.
func (c *Conn) Set(key string, value interface{}) {
	c.store.Store(key, value)
}

// Get gets the value of a key in the connection's store.
func (c *Conn) Get(key string) (interface{}, bool) {
	return c.store.Load(key)
}

// Delete deletes a key-value pair from the connection's store.
func (c *Conn) Delete(key string) {
	c.store.Delete(key)
}

// *************************
// ******** Context ********
// *************************

// Context returns the connection's context.
func (c *Conn) Context() context.Context {
	return c.context
}

// *************************
// ********* Close *********
// *************************

// Close closes the connection.
//
// It clears all event handlers and cancels the connection's context.
func (c *Conn) Close() error {
	c.eventMux.Lock()
	defer c.eventMux.Unlock()

	clear(c.events)
	c.cal()

	return c.Conn.Close()
}

func (c *Conn) newScanner() *bufio.Scanner {
	scanner := bufio.NewScanner(c.Conn)
	scanner.Split(ScanNullByte)
	return scanner
}
