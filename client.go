package gotcpio

import (
	"net"
	"sync"
)

type Client struct {
	Conn net.Conn

	// events is a map of event names to event handlers.
	events   map[string]EventHandler
	eventsMu sync.RWMutex
}

// NewClient creates a new Client.
